package etcd

import (
	"context"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"strings"
)

type Configs struct {
	db *DB
}

func (n *Configs) get(ctx context.Context, key string) (error, *cPb.Data) {
	gresp, err := n.db.Kv.Get(ctx, key)
	if err != nil {
		return err, nil
	}

	data := &cPb.Data{Data: map[string]string{}}
	for _, item := range gresp.Kvs {
		nsTask := &rPb.KV{}
		err = proto.Unmarshal(item.Value, nsTask)
		if err != nil {
			return err, nil
		}

		keyParts := strings.Split(key, "/")
		data.Data["regionid"] = keyParts[2]
		data.Data["clusterid"] = keyParts[3]
		data.Data["nodeid"] = keyParts[4]

		configs := []string{}
		for k, v := range nsTask.Extras {
			kv := strings.Join([]string{k, v.Value}, ":")
			configs = append(configs, kv)
		}
		data.Data["configs"] = strings.Join(configs, ",")
	}
	return nil, data
}

func (c *Configs) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	searchLabelsKey := helper.JoinParts("", "topology", "regions", "labels") // -> topology/regions/labels => search key
	gresp, err := c.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}

	cmp := extras["compare"]
	els := helper.SplitLabels(extras["labels"])

	datas := []*cPb.Data{}
	for _, item := range gresp.Kvs {
		newKey := helper.NewKey(string(item.Key), "configs")
		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := c.get(ctx, newKey)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := c.get(ctx, newKey)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		}
	}
	return nil, &cPb.ListResp{Data: datas}
}

// key -> topology/regions/configs/regionid/clusterid/nodes/nodeid
func (c *Configs) mutate(ctx context.Context, key, userId string, payloads []*bPb.Payload) error {
	configs := &rPb.KV{
		Extras:    map[string]*rPb.KVData{},
		Timestamp: helper.Timestamp(),
		UserId:    userId,
	}

	// Clear previous values, so that new values can take an effect
	for _, payload := range payloads {
		for pk, pv := range payload.Value {
			configs.Extras[pk] = &rPb.KVData{pv, "Waiting"}
		}
	}

	// Save node configs
	cData, err := proto.Marshal(configs)
	if err != nil {
		return err
	}

	_, err = c.db.Kv.Put(ctx, key, string(cData))
	if err != nil {
		return err
	}
	return nil
}

func (c *Configs) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	task := req.Mutate
	searchLabelsKey, kerr := helper.SearchKey(task.Task.RegionId, task.Task.ClusterId)
	if kerr != nil {
		return kerr, nil
	}
	index := []string{}

	gresp, err := c.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		newKey := helper.NewKey(string(item.Key), "configs")
		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = c.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
				index = append(index, newKey)
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = c.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
				index = append(index, newKey)
			}
		}

		// index = append(index, helper.NodeKey(string(item.Key)))
		// index = append(index, newKey)
	}

	//Save index for gravity
	req.Index = index

	// Log mutate request for resilience
	_, lerr := logMutate(ctx, req, c.db)
	if lerr != nil {
		return lerr, nil
	}

	return nil, &cPb.MutateResp{"Config added."}
}

func (c *Configs) StatusUpdate(ctx context.Context, key, newStatus string) error {
	resp, err := c.db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}

	for _, item := range resp.Kvs {
		configs := &rPb.KV{}
		err = proto.Unmarshal(item.Value, configs)
		if err != nil {
			return err
		}
		for k, _ := range configs.Extras {
			kvc := configs.Extras[k]
			configs.Extras[k] = &rPb.KVData{kvc.Value, newStatus}
		}

		// Save node configs
		cData, err := proto.Marshal(configs)
		if err != nil {
			return err
		}

		_, err = c.db.Kv.Put(ctx, key, string(cData))
		if err != nil {
			return err
		}
	}

	return nil
}
