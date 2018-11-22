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
	"strconv"
	"strings"
)

type Actions struct {
	db *DB
}

func construct(key string, nsTask *rPb.KV) *cPb.Data {
	data := &cPb.Data{Data: map[string]string{}}

	keyParts := strings.Split(key, "/")
	data.Data["regionid"] = keyParts[3]
	data.Data["clusterid"] = keyParts[4]
	data.Data["nodeid"] = keyParts[5]

	actions := []string{}
	for k, v := range nsTask.Extras {
		kv := strings.Join([]string{k, v.Value}, ":")
		actions = append(actions, kv)
	}
	actions = append(actions, strings.Join(nsTask.Index, ","))

	timestamp := helper.TSToString(nsTask.Timestamp)
	tkey := strings.Join([]string{"timestamp", timestamp}, "_")
	data.Data[tkey] = strings.Join(actions, ",")
	data.Data["index"] = strings.Join(nsTask.Index, ",")

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
		data.Data["regionid"] = keyParts[3]
		data.Data["clusterid"] = keyParts[4]
		data.Data["nodeid"] = keyParts[5]

		actions := []string{}
		for k, v := range nsTask.Extras {
			kv := strings.Join([]string{k, v.Value}, ":")
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
		timestamp := helper.TSToString(nsTask.Timestamp)
		tkey := strings.Join([]string{"timestamp", timestamp}, "_")
		data.Data[tkey] = strings.Join(actions, ",")
		data.Data["index"] = strings.Join(nsTask.Index, ",")
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
	searchLabelsKey := helper.JoinParts("", "topology", "regions", "labels") // -> topology/regions/labels => search key
	gresp, err := a.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}

	cmp := extras["compare"]
	els := helper.SplitLabels(extras["labels"])
	head, tail, from, to := getExtras(extras)

	datas := []*cPb.Data{}
	for _, item := range gresp.Kvs {
		newKey := helper.Key(string(item.Key), "actions")
		ls := helper.SplitLabels(string(item.Value))
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				gerr, data := a.get(ctx, newKey, head, tail, from, to)
				if gerr != nil {
					continue
				}
				if len(data.Data) > 0 {
					datas = append(datas, data)
				}
			}
		case "any":
			if helper.Compare(ls, els, false) {
				gerr, data := a.get(ctx, newKey, head, tail, from, to)
				if gerr != nil {
					continue
				}
				if len(data.Data) > 0 {
					datas = append(datas, data)
				}
			}
		}
	}
	return nil, &cPb.ListResp{Data: datas}
}

// key -> topology/regions/actions/regionid/clusterid/nodes/nodeid
func (a *Actions) mutate(ctx context.Context, key, userId string, payloads []*bPb.Payload) error {
	// Get what is current state of the configs for the node
	actions := &rPb.KV{
		Extras:    map[string]*rPb.KVData{},
		Timestamp: helper.Timestamp(),
		UserId:    userId,
	}

	// WHEN WORKING WITH ACTIONS WE NEED TO PRESEVER ACTIONS ORDER!
	for _, payload := range payloads {
		for _, index := range payload.Index {
			actions.Extras[index] = &rPb.KVData{payload.Value[index], "Waiting"}
		}
		actions.Index = payload.Index
	}

	// Save node actions
	aData, aerr := proto.Marshal(actions)
	if aerr != nil {
		return aerr
	}

	_, aerr = a.db.Kv.Put(ctx, key, string(aData))
	if aerr != nil {
		return aerr
	}
	return nil
}

func (a *Actions) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	// Log mutate request for resilience
	_, lerr := logMutate(ctx, req, a.db)
	if lerr != nil {
		return lerr, nil
	}

	task := req.Mutate
	searchLabelsKey, kerr := helper.SearchKey(task.Task.RegionId, task.Task.ClusterId)
	if kerr != nil {
		return kerr, nil
	}

	gresp, err := a.db.Kv.Get(ctx, searchLabelsKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		key := helper.Key(string(item.Key), "actions")
		newKey := helper.Join(key, helper.TSToString(task.Timestamp))
		ls := helper.SplitLabels(string(item.Value))
		els := helper.Labels(task.Task.Selector.Labels)

		switch task.Task.Selector.Kind {
		case bPb.CompareKind_ALL:
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				err = a.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		case bPb.CompareKind_ANY:
			if helper.Compare(ls, els, false) {
				err = a.mutate(ctx, newKey, task.UserId, task.Task.Payload)
				if err != nil {
					return err, nil
				}
			}
		}
	}
	return nil, &cPb.MutateResp{"Actions added."}
}
