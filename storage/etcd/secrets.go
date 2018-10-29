package etcd

import (
	"context"
	"errors"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"sort"
	"strings"
)

type Secrets struct {
	db *DB
}

func (n *Secrets) get(ctx context.Context, key string) (error, *cPb.Data) {
	data := &cPb.Data{Data: map[string]string{}}
	err, resp := n.db.sdb.SSecrets().List(ctx, key)
	if err != nil {
		return err, nil
	}

	keyParts := strings.Split(key, "/")
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
	cmp := extras["compare"]
	els := strings.Split(extras["labels"], ",")
	sort.Strings(els)

	datas := []*cPb.Data{}
	searchLabelsKey := helper.JoinParts("", "topology", "labels")
	gresp, err := s.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		keyPart := strings.Join(strings.Split(string(item.Key), "/labels/"), "/")
		newKey := helper.Join(keyPart, "secrets")

		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := s.get(ctx, newKey)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := s.get(ctx, newKey)
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

func conv(in map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

/*
key -> topology/regionid/clusterid/nodes/nodeid/secrets
keyPart -> topology/regionid/clusterid/nodes/nodeid to create keyPart/undone
*/
func (s *Secrets) mutate(ctx context.Context, key, keyPart string, payloads []*bPb.Payload) error {
	sData := ""
	for _, payload := range payloads {
		err, temp := s.db.sdb.SSecrets().Mutate(ctx, key, conv(payload.Value))
		if err != nil {
			return err
		}
		sData = temp
	}

	if sData != "" {
		_, serr := s.db.Kv.Put(ctx, key, sData)
		if serr != nil {
			return serr
		}

		// Add stuff to undone section
		uKey := helper.ACSUndoneKey(keyPart, "secrets")
		uresp, cerr := s.db.Kv.Get(ctx, uKey)
		if cerr != nil {
			return cerr
		}

		if len(uresp.Kvs) == 0 {
			extras := map[string]string{sData: ""}
			undone := &rPb.KV{Extras: extras}
			uData, uerr := proto.Marshal(undone)
			if uerr != nil {
				return uerr
			}

			_, cerr = s.db.Kv.Put(ctx, uKey, string(uData))
			if cerr != nil {
				return cerr
			}
		}
	}

	return nil
}

func (c *Secrets) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	task := req.Mutate
	searchLabelsKey := ""
	if task.Task.RegionId == "*" && task.Task.ClusterId == "*" {
		searchLabelsKey = helper.JoinParts("", "topology", "labels") // topology/labels/
	} else if task.Task.RegionId != "*" && task.Task.ClusterId == "*" {
		searchLabelsKey = helper.JoinParts("", "topology", "labels", task.Task.RegionId) //topology/labels/regionid/
	} else if task.Task.RegionId != "*" && task.Task.ClusterId != "*" { //topology/labels/regionid/clusterid/
		searchLabelsKey = helper.JoinParts("", "topology", "labels", task.Task.RegionId, task.Task.ClusterId)
	} else {
		return errors.New("Request not valid"), nil
	}

	gresp, err := c.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		keyPart := strings.Join(strings.Split(string(item.Key), "/labels/"), "/")
		newKey := helper.Join(keyPart, "secrets")

		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = c.mutate(ctx, newKey, keyPart, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = c.mutate(ctx, newKey, keyPart, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		}
	}
	return nil, &cPb.MutateResp{"Secrets added."}
}
