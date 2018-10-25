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
	"strings"
)

type Actions struct {
	db *DB
}

func (a *Actions) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	return nil, nil
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
	for _, payload := range payloads {
		for pk, pv := range payload.Value {
			actions.Extras[pk] = pv
		}
	}
	actions.Timestamp = helper.Timestamp() // add a timestamp. Actoins are grouped byt time!

	// Save node configs
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
		}

		// update undone with new stuff thats been added/removed/upddated
		for k, v := range actions.Extras {
			undone.Undone["actions"].Extras[k] = v
		}

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
