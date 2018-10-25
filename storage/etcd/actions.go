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

type Actions struct {
	db *DB
}

func (a *Actions) get(ctx context.Context, key string) (error, *cPb.Data) {
	data := &cPb.Data{Data: map[string]string{}}
	gresp, err := a.db.Kv.Get(ctx, key)
	if err != nil {
		return err, nil
	}

	for _, item := range gresp.Kvs {
		nsTask := &rPb.KV{}
		err = proto.Unmarshal(item.Value, nsTask)
		if err != nil {
			return err, nil
		}

		keyParts := strings.Split(key, "/")
		data.Data["regionid"] = keyParts[1]
		data.Data["clusterid"] = keyParts[2]
		data.Data["nodeid"] = keyParts[3]

		configs := []string{}
		for k, v := range nsTask.Extras {
			kv := strings.Join([]string{k, v}, ":")
			configs = append(configs, kv)
		}
		data.Data["configs"] = strings.Join(configs, ",")

		return nil, data
	}
	return err, nil
}

func (a *Actions) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	cmp := extras["compare"]
	els := strings.Split(extras["labels"], ",")
	sort.Strings(els)

	// top := 0
	// from := 0
	// to := 0

	datas := []*cPb.Data{}
	searchLabelsKey := helper.JoinParts("", "topology", "labels")
	gresp, err := a.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		keyPart := strings.Join(strings.Split(string(item.Key), "/labels/"), "/")
		newKey := helper.Join(keyPart, "actions")

		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := a.get(ctx, newKey)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := a.get(ctx, newKey)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		}
	}
	return nil, &cPb.ListResp{Data: datas}
}

/*
key -> topology/regionid/clusterid/nodes/nodeid/configs
keyPart -> topology/regionid/clusterid/nodes/nodeid to create keyPart/undone
*/
func (a *Actions) mutate(ctx context.Context, key, keyPart string, payloads []*bPb.Payload) error {
	cresp, cerr := a.db.Kv.Get(ctx, key)
	if cerr != nil {
		return cerr
	}

	// Get what is current state of the configs for the node
	actions := &rPb.KV{}
	for _, citem := range cresp.Kvs {
		cerr = proto.Unmarshal(citem.Value, actions)
		if cerr != nil {
			return cerr
		}
	}
	if actions.Extras == nil {
		actions.Extras = map[string]string{}
	}

	// Than test if submited configs exits or not.
	// If exists and value is Tombstone than mark that configs for deletation
	// and executor service will remove it from the nodes
	// If exists and value is not Tombstone than replace value with new value
	// WHEN WORKING WITH ACTIONS WE NEED TO PRESEVER ACTIONS ORDER!
	for _, payload := range payloads {
		for _, index := range payload.Index {
			actions.Extras[index] = payload.Value[index]
		}
		actions.Index = payload.Index
	}
	actions.Timestamp = helper.Timestamp() // add a timestamp. Actoins are grouped by time!

	// Save node actions
	aData, aerr := proto.Marshal(actions)
	if aerr != nil {
		return aerr
	}

	_, aerr = a.db.Kv.Put(ctx, key, string(aData))
	if aerr != nil {
		return aerr
	}

	// Add stuff to undone section
	uKey := helper.ACSUndoneKey(keyPart)
	uresp, cerr := a.db.Kv.Get(ctx, uKey)
	if cerr != nil {
		return cerr
	}

	for _, uitem := range uresp.Kvs {
		undone := &rPb.UndoneKV{}
		cerr = proto.Unmarshal(uitem.Value, undone)
		if cerr != nil {
			return cerr
		}

		if undone.Undone == nil {
			undone.Undone = map[string]*rPb.KV{}
			undone.Undone["actions"] = &rPb.KV{Extras: map[string]string{}}
		} else if _, ok := undone.Undone["actions"]; !ok {
			undone.Undone["actions"] = &rPb.KV{Extras: map[string]string{}}
		}

		// update undone with new stuff thats been added/removed/upddated
		for k, v := range actions.Extras {
			undone.Undone["actions"].Extras[k] = v
		}

		//save index
		undone.Undone["actions"].Index = actions.Index

		uData, uerr := proto.Marshal(undone)
		if uerr != nil {
			return uerr
		}

		_, cerr = a.db.Kv.Put(ctx, uKey, string(uData))
		if cerr != nil {
			return cerr
		}
	}
	return nil
}

func (a *Actions) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
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

	gresp, err := a.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		keyPart := strings.Join(strings.Split(string(item.Key), "/labels/"), "/")
		newKey := helper.Join(keyPart, "actions")

		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = a.mutate(ctx, newKey, keyPart, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = a.mutate(ctx, newKey, keyPart, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		}
	}
	return nil, &cPb.MutateResp{"Actions added."}
}
