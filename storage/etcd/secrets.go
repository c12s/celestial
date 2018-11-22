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

type Secrets struct {
	db *DB
}

func (n *Secrets) get(ctx context.Context, key, user string) (error, *cPb.Data) {
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

		keyParts := strings.Split(string(item.Key), "/")
		data.Data["regionid"] = keyParts[2]
		data.Data["clusterid"] = keyParts[3]
		data.Data["nodeid"] = keyParts[4]

		err, resp := n.db.sdb.SSecrets().List(ctx, string(item.Key), user)
		if err != nil {
			return err, nil
		}

		secrets := []string{}
		for k, v := range resp {
			kv := strings.Join([]string{k, v}, ":")
			secrets = append(secrets, kv)
		}
		data.Data["secrets"] = strings.Join(secrets, ",")
	}
	return nil, data
}

func (s *Secrets) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	searchLabelsKey := helper.JoinParts("", "topology", "regions", "labels") // -> topology/regions/labels => search key
	gresp, err := s.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}

	cmp := extras["compare"]
	els := helper.SplitLabels(extras["labels"])
	userId := extras["user"]

	datas := []*cPb.Data{}
	for _, item := range gresp.Kvs {
		newKey := helper.NewKey(string(item.Key), "secrets")
		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := s.get(ctx, newKey, userId)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := s.get(ctx, newKey, userId)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		}
	}
	return nil, &cPb.ListResp{Data: datas}
}

func toSecrets(payloads []*bPb.Payload) map[string]interface{} {
	input := map[string]interface{}{}
	for _, payload := range payloads {
		for pk, pv := range payload.Value {
			input[pk] = pv
		}
	}
	return input
}

// key -> topology/regions/secrets/regionid/clusterid/nodes/nodeid
func (s *Secrets) mutate(ctx context.Context, key, userId string, payloads []*bPb.Payload) error {
	temp := toSecrets(payloads)
	err, sData := s.db.sdb.SSecrets().Mutate(ctx, key, userId, temp)
	if err != nil {
		return err
	}

	if sData != "" {
		secrets := &rPb.KV{
			Extras:    map[string]*rPb.KVData{},
			Timestamp: helper.Timestamp(),
			UserId:    userId,
		}
		for k, _ := range temp {
			secrets.Extras[k] = &rPb.KVData{sData, "Waiting"}
		}

		// Save node secrets in kv store just path to secrets store
		data, serr := proto.Marshal(secrets)
		if serr != nil {
			return serr
		}

		_, serr = s.db.Kv.Put(ctx, key, string(data))
		if serr != nil {
			return serr
		}
	}
	return nil
}

func (c *Secrets) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	// Log mutate request for resilience
	_, lerr := logMutate(ctx, req, c.db)
	if lerr != nil {
		return lerr, nil
	}

	task := req.Mutate
	searchLabelsKey, kerr := helper.SearchKey(task.Task.RegionId, task.Task.ClusterId)
	if kerr != nil {
		return kerr, nil
	}

	gresp, err := c.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		newKey := helper.NewKey(string(item.Key), "secrets")
		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = c.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = c.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		}
	}
	return nil, &cPb.MutateResp{"Secrets added."}
}
