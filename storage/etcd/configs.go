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
	"strconv"
	"strings"
)

const (
	Tombstone = "!!Tombstone"
)

type Configs struct {
	db *DB
}

func (n *Configs) get(ctx context.Context, key string) (string, int64, string, string) {
	gresp, err := n.db.Kv.Get(ctx, key)
	if err != nil {
		return "", 0, "", ""
	}

	for _, item := range gresp.Kvs {
		nsTask := &rPb.Task{}
		err = proto.Unmarshal(item.Value, nsTask)
		if err != nil {
			return "", 0, "", ""
		}
		return nsTask.Namespace, nsTask.Timestamp, nsTask.Extras["namespace"], nsTask.Extras["labels"]
	}
	return "", 0, "", ""
}

func (c *Configs) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	cmp := extras["compare"]
	els := strings.Split(extras["labels"], ",")
	sort.Strings(els)

	datas := []*cPb.Data{}
	gresp, err := c.db.Kv.Get(ctx, helper.NSLabels(), clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err, nil
	}
	for _, item := range gresp.Kvs {
		key := string(item.Key)
		newKey := strings.Join(strings.Split(key, "/labels/"), "/")
		ls := strings.Split(string(item.Value), ",")
		sort.Strings(ls)

		data := &cPb.Data{Data: map[string]string{}}
		switch cmp {
		case "all":
			if len(ls) == len(els) && helper.Compare(ls, els, true) {
				ns, timestamp, name, labels := c.get(ctx, newKey)
				if ns != "" {
					data.Data["namespace"] = ns
					data.Data["age"] = strconv.FormatInt(timestamp, 10)
					data.Data["name"] = name
					data.Data["labels"] = labels
				}
				datas = append(datas, data)
			}
		case "any":
			if helper.Compare(ls, els, false) {
				ns, timestamp, name, labels := c.get(ctx, newKey)
				if ns != "" {
					data.Data["namespace"] = ns
					data.Data["age"] = strconv.FormatInt(timestamp, 10)
					data.Data["name"] = name
					data.Data["labels"] = labels
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
func (c *Configs) mutate(ctx context.Context, key, keyPart string, payloads []*bPb.Payload) error {
	cresp, cerr := c.db.Kv.Get(ctx, key)
	if cerr != nil {
		return cerr
	}

	// Get what is current state of the configs for the node
	configs := &rPb.KV{}
	for _, citem := range cresp.Kvs {
		cerr = proto.Unmarshal(citem.Value, configs)
		if cerr != nil {
			return cerr
		}
	}
	if configs.Extras == nil {
		configs.Extras = map[string]string{}
	}

	// Than test if submited configs exits or not.
	// If exists and value is Tombstone than mark that configs for deletation
	// and executor service will remove it from the nodes
	// If exists and value is not Tombstone than replace value with new value
	for _, payload := range payloads {
		for pk, pv := range payload.Value {
			if _, ok := configs.Extras[pk]; ok {
				if pv == Tombstone {
					configs.Extras[pk] = Tombstone
				} else {
					configs.Extras[pk] = pv
				}
			} else {
				configs.Extras[pk] = pv
			}
		}
	}

	// Save node configs
	cData, serr := proto.Marshal(configs)
	if serr != nil {
		return serr
	}

	_, cerr = c.db.Kv.Put(ctx, key, string(cData))
	if cerr != nil {
		return cerr
	}

	// Add stuff to undone section
	uKey := helper.ACSUndoneKey(keyPart)
	uresp, cerr := c.db.Kv.Get(ctx, uKey)
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
			undone.Undone["configs"] = &rPb.KV{Extras: map[string]string{}}
		}

		// update undone with new stuff thats been added/removed/upddated
		for k, v := range configs.Extras {
			undone.Undone["configs"].Extras[k] = v
		}

		uData, uerr := proto.Marshal(undone)
		if uerr != nil {
			return uerr
		}

		_, cerr = c.db.Kv.Put(ctx, uKey, string(uData))
		if cerr != nil {
			return cerr
		}
	}

	return nil
}

func (c *Configs) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
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
		newKey := helper.Join(keyPart, "configs")

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
	return nil, &cPb.MutateResp{"Config added."}
}
