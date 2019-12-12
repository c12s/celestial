package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/helper"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"sort"
	"strconv"
	"strings"
)

type Namespaces struct {
	db *DB
}

func (n *Namespaces) get(ctx context.Context, key string) (string, int64, string, string) {
	span, _ := sg.FromGRPCContext(ctx, "get")
	defer span.Finish()
	fmt.Println(span)

	chspan := span.Child("etcd.get")
	gresp, err := n.db.Kv.Get(ctx, key)
	if err != nil {
		chspan.AddLog(&sg.KV{"etcd get error", err.Error()})
		return "", 0, "", ""
	}
	go chspan.Finish()

	for _, item := range gresp.Kvs {
		nsTask := &rPb.Task{}
		err = proto.Unmarshal(item.Value, nsTask)
		if err != nil {
			span.AddLog(&sg.KV{"unmarshall etcd get error", err.Error()})
			return "", 0, "", ""
		}
		return nsTask.Namespace, nsTask.Timestamp, nsTask.Extras["namespace"], nsTask.Extras["labels"]
	}
	return "", 0, "", ""
}

func (n *Namespaces) List(ctx context.Context, extra map[string]string) (error, *cPb.ListResp) {
	span, _ := sg.FromGRPCContext(ctx, "list")
	defer span.Finish()
	fmt.Println(span)

	name := extra["name"]
	cmp := extra["compare"]
	els := strings.Split(extra["labels"], ",")
	sort.Strings(els)

	datas := []*cPb.Data{}
	if name == "" {
		chspan := span.Child("etcd.get searchLabels")
		gresp, err := n.db.Kv.Get(ctx, helper.NSLabels(), clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
		if err != nil {
			chspan.AddLog(&sg.KV{"etcd.get error", err.Error()})
			return err, nil
		}
		go chspan.Finish()

		for _, item := range gresp.Kvs {
			key := string(item.Key)
			newKey := strings.Join(strings.Split(key, "/labels/"), "/")
			ls := helper.SplitLabels(string(item.Value))

			data := &cPb.Data{Data: map[string]string{}}
			switch cmp {
			case "all":
				if len(ls) == len(els) && helper.Compare(ls, els, true) {
					ns, timestamp, name, labels := n.get(sg.NewTracedGRPCContext(ctx, span), newKey)
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
					ns, timestamp, name, labels := n.get(sg.NewTracedGRPCContext(ctx, span), newKey)
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
	} else {
		data := &cPb.Data{Data: map[string]string{}}
		ns, timestamp, name, labels := n.get(sg.NewTracedGRPCContext(ctx, span), helper.NSKey(name))
		if ns != "" {
			data.Data["namespace"] = ns
			data.Data["age"] = strconv.FormatInt(timestamp, 10)
			data.Data["name"] = name
			data.Data["labels"] = labels
		}
		datas = append(datas, data)
	}
	return nil, &cPb.ListResp{Data: datas}
}

func (n *Namespaces) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	task := req.Mutate
	namespace := task.Extras["namespace"]
	labels := task.Extras["labels"]

	nsKey := helper.NSKey(namespace)
	nsData, _ := proto.Marshal(task)

	chspan1 := span.Child("etcd.put key")
	_, err := n.db.Kv.Put(ctx, nsKey, string(nsData))
	if err != nil {
		chspan1.AddLog(&sg.KV{"etcd.put key error", err.Error()})
		return err, nil
	}
	chspan1.Finish()

	chspan2 := span.Child("etcd.put labels")
	lKey := helper.NSLabelsKey(namespace)
	_, err = n.db.Kv.Put(ctx, lKey, labels)
	if err != nil {
		chspan2.AddLog(&sg.KV{"etcd.put labels error", err.Error()})
		return err, nil
	}
	chspan2.Finish()

	chspan3 := span.Child("etcd.put status")
	sKey := helper.NSStatusKey(namespace)
	_, err = n.db.Kv.Put(ctx, sKey, "Done")
	if err != nil {
		chspan3.AddLog(&sg.KV{"etcd.put status error", err.Error()})
		return err, nil
	}
	chspan3.Finish()

	return nil, &cPb.MutateResp{"Namespaces added."}
}
