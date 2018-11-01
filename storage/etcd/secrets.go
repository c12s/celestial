package etcd

import (
	"context"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
)

type Secrets struct {
	db *DB
}

func (n *Secrets) get(ctx context.Context, key, user string) (error, *cPb.Data) {
	err, resp := n.db.sdb.SSecrets().List(ctx, key, user)
	if err != nil {
		return err, nil
	}

	keyParts := strings.Split(key, "/")
	data := &cPb.Data{Data: map[string]string{}}
	data.Data["regionid"] = keyParts[1]
	data.Data["clusterid"] = keyParts[2]
	data.Data["nodeid"] = keyParts[3]

	secrets := []string{}
	for k, v := range resp {
		kv := strings.Join([]string{k, v}, ":")
		secrets = append(secrets, kv)
	}
	data.Data["secrets"] = strings.Join(secrets, ",")
	return err, data
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
		newKey := helper.Key(string(item.Key), "secrets")
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

	return nil, nil
}

// key -> topology/regions/secrets/regionid/clusterid/nodes/nodeid
func (s *Secrets) mutate(ctx context.Context, key, userId string, payloads []*bPb.Payload) error {
	sresp, serr := s.db.Kv.Get(ctx, key)
	if serr != nil {
		return serr
	}

	// Get what is current state of the configs for the node
	secrets := &rPb.KV{}
	for _, sitem := range sresp.Kvs {
		serr = proto.Unmarshal(sitem.Value, secrets)
		if serr != nil {
			return serr
		}
	}
	if secrets.Extras == nil {
		secrets.Extras = map[string]*rPb.KVData{}
	}

	input := map[string]interface{}{}
	for _, payload := range payloads {
		for pk, pv := range payload.Value {
			if _, ok := secrets.Extras[pk]; ok {
				if pv == Tombstone {
					secrets.Extras[pk] = &rPb.KVData{Tombstone, "Waiting"}
				} else {
					secrets.Extras[pk] = &rPb.KVData{"", "Waiting"}
				}
			} else {
				secrets.Extras[pk] = &rPb.KVData{"", "Waiting"}
			}
			input[pk] = pv
		}
	}
	secrets.Timestamp = helper.Timestamp()
	secrets.UserId = userId

	err, sData := s.db.sdb.SSecrets().Mutate(ctx, key, userId, input)
	if err != nil {
		return err
	}

	if sData != "" {
		for k, _ := range secrets.Extras {
			secrets.Extras[k] = &rPb.KVData{sData, "Waiting"}
		}

		// Save node configs
		sData, sserr := proto.Marshal(secrets)
		if sserr != nil {
			return sserr
		}

		_, serr = s.db.Kv.Put(ctx, key, string(sData))
		if serr != nil {
			return serr
		}
	}

	return nil
}

func (c *Secrets) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
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
		key := helper.Key(string(item.Key), "secrets")
		newKey := helper.Join(key, strconv.FormatInt(task.Timestamp, 10))
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
