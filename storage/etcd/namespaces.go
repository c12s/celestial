package etcd

import (
	"context"
	"github.com/c12s/celestial/helper"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"sort"
	"strconv"
	"strings"
)

type Namespaces struct {
	db *DB
}

func compare(a, b []string, strict bool) bool {
	for _, akv := range a {
		for _, bkv := range b {
			if akv == bkv && !strict {
				return true
			}
		}
	}
	return true
}

func (n *Namespaces) get(ctx context.Context, key string) (string, int64, string, string) {
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

func (n *Namespaces) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	result := &cPb.ListResp{Data: map[string]string{}}
	name := extras["name"]
	cmp := extras["compare"]
	els := strings.Split(extras["labels"], ",")
	sort.Strings(els)
	if name == "" {
		gresp, err := n.db.Kv.Get(ctx, helper.Labels(), clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
		if err != nil {
			return err, nil
		}
		for _, item := range gresp.Kvs {
			key := string(item.Key)
			newKey := strings.Join(strings.Split(key, "/labels/"), "/")
			ls := strings.Split(string(item.Value), ",")
			sort.Strings(ls)
			switch cmp {
			case "all":
				if len(ls) == len(els) && compare(ls, els, true) {
					ns, timestamp, name, labels := n.get(ctx, newKey)
					if ns != "" {
						result.Data["namespace"] = ns
						result.Data["age"] = strconv.FormatInt(timestamp, 10)
						result.Data["name"] = name
						result.Data["labels"] = labels
					}
				}
			case "any":
				if compare(ls, els, false) {
					ns, timestamp, name, labels := n.get(ctx, newKey)
					if ns != "" {
						result.Data["namespace"] = ns
						result.Data["age"] = strconv.FormatInt(timestamp, 10)
						result.Data["name"] = name
						result.Data["labels"] = labels
					}
				}
			}
		}
	} else {
		ns, timestamp, name, labels := n.get(ctx, helper.NSKey(name))
		if ns != "" {
			result.Data["namespace"] = ns
			result.Data["age"] = strconv.FormatInt(timestamp, 10)
			result.Data["name"] = name
			result.Data["labels"] = labels
		}
	}
	return nil, result
}

func (n *Namespaces) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	task := req.Mutate
	namespace := task.Extras["namespace"]
	labels := task.Extras["labels"]

	nsKey := helper.NSKey(namespace)
	nsData, _ := proto.Marshal(task)
	_, err := n.db.Kv.Put(ctx, nsKey, string(nsData))
	if err != nil {
		return err, nil
	}

	lKey := helper.NSLabelsKey(namespace)
	_, err = n.db.Kv.Put(ctx, lKey, labels)
	if err != nil {
		return err, nil
	}

	sKey := helper.NSStatusKey(namespace)
	_, err = n.db.Kv.Put(ctx, sKey, "Active")
	if err != nil {
		return err, nil
	}

	return nil, &cPb.MutateResp{"Namespaces added."}
}
