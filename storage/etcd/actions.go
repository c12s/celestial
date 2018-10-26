package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"sort"
	"strconv"
	"strings"
)

type Actions struct {
	db *DB
}

func construct(key string, nsTask *rPb.KV) *cPb.Data {
	data := &cPb.Data{Data: map[string]string{}}

	keyParts := strings.Split(key, "/")
	data.Data["regionid"] = keyParts[1]
	data.Data["clusterid"] = keyParts[2]
	data.Data["nodeid"] = keyParts[3]

	actions := []string{}
	for k, v := range nsTask.Extras {
		kv := strings.Join([]string{k, v}, ":")
		actions = append(actions, kv)
	}
	data.Data["actions"] = strings.Join(actions, ",")
	data.Data["timestamp"] = strconv.FormatInt(nsTask.Timestamp, 10)

	return data
}

func (a *Actions) getHT(ctx context.Context, key string, head, tail int64) (error, *cPb.Data) {
	if head != 0 {
		gresp, err := a.db.Kv.Get(ctx, key, clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend), clientv3.WithLimit(head))
		if err != nil {
			return err, nil
		}
		for _, item := range gresp.Kvs {
			nsTask := &rPb.KV{}
			err = proto.Unmarshal(item.Value, nsTask)
			if err != nil {
				return err, nil
			}
			return nil, construct(key, nsTask)
		}
	} else if tail != 0 {
		gresp, err := a.db.Kv.Get(ctx, key, clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend), clientv3.WithLimit(tail))
		if err != nil {
			return err, nil
		}
		for _, item := range gresp.Kvs {
			nsTask := &rPb.KV{}
			err = proto.Unmarshal(item.Value, nsTask)
			if err != nil {
				return err, nil
			}
			return nil, construct(key, nsTask)
		}
	}
	return errors.New("Cant use hand and tail at the same time!"), nil
}

func (a *Actions) getFT(ctx context.Context, key string, from, to int64) (error, *cPb.Data) {
	data := &cPb.Data{Data: map[string]string{}}
	gresp, err := a.db.Kv.Get(ctx, key,
		clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
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

		actions := []string{}
		for k, v := range nsTask.Extras {
			kv := strings.Join([]string{k, v}, ":")
			if from != 0 && to != 0 {
				if nsTask.Timestamp >= from && nsTask.Timestamp <= to {
					actions = append(actions, kv)
				}
			} else if from != 0 && to == 0 {
				if nsTask.Timestamp >= from {
					actions = append(actions, kv)
				}
			} else if from == 0 && to != 0 {
				if nsTask.Timestamp <= to {
					actions = append(actions, kv)
				}
			} else {
				actions = append(actions, kv)
			}
		}
		if val, ok := data.Data["actions"]; ok {
			newVal := strings.Join(actions, ",")
			data.Data["actions"] = strings.Join([]string{val, newVal}, ",")
		} else {
			data.Data["actions"] = strings.Join(actions, ",")
		}
		data.Data["timestamp"] = strconv.FormatInt(nsTask.Timestamp, 10)
	}
	return nil, data
}

func getExtras(extras map[string]string) (int64, int64, int64, int64) {
	head := int64(0)
	tail := int64(0)
	from := int64(0)
	to := int64(0)

	if val, ok := extras["head"]; ok {
		head, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := extras["tail"]; ok {
		tail, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := extras["from"]; ok {
		from, _ = strconv.ParseInt(val, 10, 64)
	}

	if val, ok := extras["to"]; ok {
		to, _ = strconv.ParseInt(val, 10, 64)
	}

	return head, tail, from, to
}

func (a *Actions) get(ctx context.Context, key string, head, tail, from, to int64) (error, *cPb.Data) {
	if head != 0 || tail != 0 {
		return a.getHT(ctx, key, head, tail)
	}
	return a.getFT(ctx, key, from, to)
}

func (a *Actions) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	cmp := extras["compare"]
	els := strings.Split(extras["labels"], ",")
	sort.Strings(els)
	head, tail, from, to := getExtras(extras)

	datas := []*cPb.Data{}
	searchLabelsKey := helper.JoinParts("", "topology", "labels")
	gresp, err := a.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		fmt.Println("gresp err ", err)
		return err, nil
	}
	for _, item := range gresp.Kvs {
		keyPart := strings.Join(strings.Split(string(item.Key), "/labels/"), "/")
		newKey := helper.Join(keyPart, "actions")

		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := a.get(ctx, newKey, head, tail, from, to)
				if gerr != nil {
					continue
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := a.get(ctx, newKey, head, tail, from, to)
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
		newKey := helper.JoinFull(keyPart, "actions", strconv.FormatInt(task.Timestamp, 10)) // add timestamp at the end of the key, actions are grouped by time

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
