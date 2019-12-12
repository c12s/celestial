package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"strings"
)

type Secrets struct {
	db *DB
}

func (n *Secrets) get(ctx context.Context, key, user string) (error, *cPb.Data) {
	span, _ := sg.FromGRPCContext(ctx, "get")
	defer span.Finish()
	fmt.Println(span)

	chspan := span.Child("etcd.get")
	gresp, err := n.db.Kv.Get(ctx, key)
	if err != nil {
		chspan.AddLog(&sg.KV{"etcd get error", err.Error()})
		return err, nil
	}
	go chspan.Finish()

	data := &cPb.Data{Data: map[string]string{}}
	for _, item := range gresp.Kvs {
		nsTask := &rPb.KV{}
		err = proto.Unmarshal(item.Value, nsTask)
		if err != nil {
			span.AddLog(&sg.KV{"unmarshall etcd get error", err.Error()})
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
	span, _ := sg.FromGRPCContext(ctx, "list")
	defer span.Finish()
	fmt.Println(span)

	chspan := span.Child("etcd.get searchLabels")
	searchLabelsKey := helper.JoinParts("", "topology", "regions", "labels") // -> topology/regions/labels => search key
	gresp, err := s.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		chspan.AddLog(&sg.KV{"etcd.get error", err.Error()})
		return err, nil
	}
	go chspan.Finish()

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
				gerr, data := s.get(sg.NewTracedGRPCContext(ctx, span), newKey, userId)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := s.get(sg.NewTracedGRPCContext(ctx, span), newKey, userId)
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
	span, _ := sg.FromGRPCContext(ctx, "helper mutate")
	defer span.Finish()
	fmt.Println(span)

	temp := toSecrets(payloads)
	err, sData := s.db.sdb.SSecrets().Mutate(sg.NewTracedGRPCContext(ctx, span), key, userId, temp)
	if err != nil {
		span.AddLog(&sg.KV{"secrets db mutate error", err.Error()})
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
			span.AddLog(&sg.KV{"marshaling error", err.Error()})
			return serr
		}

		chspan := span.Child("etcd.put")
		_, serr = s.db.Kv.Put(ctx, key, string(data))
		if serr != nil {
			chspan.AddLog(&sg.KV{"etcd.put error", err.Error()})
			return serr
		}
		chspan.Finish()
	}
	return nil
}

func (c *Secrets) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	task := req.Mutate
	searchLabelsKey, kerr := helper.SearchKey(task.Task.RegionId, task.Task.ClusterId)
	if kerr != nil {
		span.AddLog(&sg.KV{"search key error", kerr.Error()})
		return kerr, nil
	}
	index := []string{}

	chspan := span.Child("etcd.get searchLabels")
	gresp, err := c.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		chspan.AddLog(&sg.KV{"etcd.get error", err.Error()})
		return err, nil
	}
	go chspan.Finish()

	for _, item := range gresp.Kvs {
		newKey := helper.NewKey(string(item.Key), "secrets")
		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = c.mutate(sg.NewTracedGRPCContext(ctx, span), newKey, task.UserId, task.Task.Payload)
				if err != nil {
					span.AddLog(&sg.KV{"mutate error", err.Error()})
					return err, nil
				}
				index = append(index, newKey)
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = c.mutate(sg.NewTracedGRPCContext(ctx, span), newKey, task.UserId, task.Task.Payload)
				if err != nil {
					span.AddLog(&sg.KV{"mutate error", err.Error()})
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
	_, lerr := logMutate(sg.NewTracedGRPCContext(ctx, span), req, c.db)
	if lerr != nil {
		span.AddLog(&sg.KV{"resilience log error", lerr.Error()})
		return lerr, nil
	}

	span.AddLog(&sg.KV{"config addition", "Config added."})
	return nil, &cPb.MutateResp{"Secrets added."}
}

func (s *Secrets) StatusUpdate(ctx context.Context, key, newStatus string) error {
	resp, err := s.db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}

	for _, item := range resp.Kvs {
		secrets := &rPb.KV{}
		err = proto.Unmarshal(item.Value, secrets)
		if err != nil {
			return err
		}
		for k, _ := range secrets.Extras {
			kvc := secrets.Extras[k]
			secrets.Extras[k] = &rPb.KVData{kvc.Value, newStatus}
		}

		// Save node configs
		cData, err := proto.Marshal(secrets)
		if err != nil {
			return err
		}

		_, err = s.db.Kv.Put(ctx, key, string(cData))
		if err != nil {
			return err
		}
	}

	return nil
}
