package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/service"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	gPb "github.com/c12s/scheme/gravity"
	mPb "github.com/c12s/scheme/magnetar"
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
)

type Topology struct {
	db *DB
}

func (t *Topology) StatusUpdate(ctx context.Context, key, newStatus string) error {
	return nil
}

func (t *Topology) Watcher(ctx context.Context) {
	go func(c context.Context) {
		// When service is started, launch saved watchers
		resp, err := t.db.Kv.Get(ctx, "watcher/", clientv3.WithPrefix())
		if err != nil {
			fmt.Println(err.Error())
		}

		for _, ev := range resp.Kvs {
			t.watch(ctx, string(ev.Value))
		}

		// And start listerner for new created
		rch := t.db.Client.Watch(c, "watcher/", clientv3.WithPrefix())
		for {
			select {
			case result := <-rch:
				for _, ev := range result.Events {
					if ev.Type == clientv3.EventTypePut {
						fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
						t.watch(ctx, string(ev.Kv.Value))
					}
				}
			case <-c.Done():
				return
			}
		}
	}(ctx)
}

func (t *Topology) watch(ctx context.Context, key string) {
	go func(c context.Context) {
		rch := t.db.Client.Watch(c, key, clientv3.WithPrefix())
		for {
			select {
			case result := <-rch:
				for _, ev := range result.Events {
					if ev.Type == clientv3.EventTypePut {
						// TODO: For now just update status, as is but in future
						// Since this will store machine info we should get whole data
						// And update not just status but all machine data!!!
						_, err := t.db.Kv.Put(ctx, string(ev.Kv.Key), "Alive")
						if err != nil {
							fmt.Println(err.Error())
						}
					} else if ev.Type == clientv3.EventTypeDelete {
						// TODO: For now just update status, as is but in future
						// Since this will store machine info we should get whole data
						// And update just status!!!
						_, err := t.db.Kv.Put(ctx, string(ev.Kv.Key), "Down")
						if err != nil {
							fmt.Println(err.Error())
						}
					}
				}
			case <-c.Done():
				return
			}

		}
	}(ctx)
}

func (t *Topology) mutate(ctx context.Context, availableNodes []string, tmpLabels map[string]string, userid, namespace string) error {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	for _, id := range availableNodes {
		rid, cid, nid := helper.GetParts(id)
		keys := helper.TKeys(rid, cid, nid, userid, namespace)

		cData, err := proto.Marshal(&rPb.KV{})
		if err != nil {
			span.AddLog(
				&sg.KV{"Marshaling error", err.Error()},
			)
		}
		ops := []clientv3.Op{
			clientv3.OpPut(keys["labels"], tmpLabels[id]),
			clientv3.OpPut(keys["nodeid"], "Pending"),
			clientv3.OpPut(keys["configs"], string(cData)),
			clientv3.OpPut(keys["secrets"], string(cData)),
			clientv3.OpPut(keys["actions"], string(cData)),
		}
		for _, op := range ops {
			chspan := span.Child("etcd.do")
			if _, err := t.db.Kv.Do(ctx, op); err != nil {
				chspan.AddLog(
					&sg.KV{"Marshaling error", err.Error()},
				)
			}
			chspan.Finish()
		}

		//start watcher for prefix topology/regions/clusterid
		// so there is one watcher per cluster, and nodes
		// will send one watcher notified
		_, err = t.db.Kv.Put(ctx, helper.WatcherKey(keys["watch"]), keys["watch"])
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Topology) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	span, _ := sg.FromGRPCContext(ctx, "mutate")
	defer span.Finish()
	fmt.Println(span)

	// Log mutate request for resilience
	logTaskKey, lerr := logMutate(sg.NewTracedGRPCContext(ctx, span), req, t.db)
	if lerr != nil {
		span.AddLog(&sg.KV{"resilience log error", lerr.Error()})
		return lerr, nil
	}

	ids := []string{}
	tk := req.Mutate.Task
	tmpLabels := map[string]string{}
	for _, p := range tk.Payload {
		id := helper.ReserveKey(tk.RegionId, tk.ClusterId, p.Value["ID"])
		p.Value["NEW_ID"] = newId(id)
		ids = append(ids, id)
		tmpLabels[id] = tolabels(p.Value)
	}

	client := service.NewMagnetarClient(t.db.Magnetar)
	resp, err := client.Reserve(sg.NewTracedGRPCContext(ctx, span), &mPb.ReserveMsg{Ids: ids})
	if err != nil {
		span.AddLog(
			&sg.KV{"Magnetar error", err.Error()},
		)
		return err, nil
	}

	if int32(len(ids)) == resp.Taken {
		span.AddLog(
			&sg.KV{"Reservation", "All specified nodes reserved"},
		)
	}

	availableNodes := difference(ids, resp.Excluded)
	err = t.mutate(ctx, availableNodes, tmpLabels, req.Mutate.UserId, req.Mutate.Namespace)
	if err != nil {
		return err, nil
	}

	//Save index for gravity
	req.Index = index(availableNodes)
	err = t.sendToGravity(sg.NewTracedGRPCContext(ctx, span), req, logTaskKey)
	if err != nil {
		span.AddLog(&sg.KV{"putTask error", err.Error()})
		return err, nil
	}

	fmt.Println("{{RESERVED}}: ", resp.Taken)
	return nil, &cPb.MutateResp{Error: "Some nodes reserved, check stellar."}
}

func (c *Topology) sendToGravity(ctx context.Context, req *cPb.MutateReq, taskKey string) error {
	span, _ := sg.FromGRPCContext(ctx, "sendToGravity")
	defer span.Finish()
	fmt.Println(span)
	fmt.Println("SERIALIZE ", span.Serialize())

	token, err := helper.ExtractToken(ctx)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return err
	}

	client := service.NewGravityClient(c.db.Gravity)
	gReq := &gPb.PutReq{
		Key:     taskKey, //key to be deleted after push is done
		Task:    req,
		TaskKey: taskKey,
	}

	_, err = client.PutTask(
		helper.AppendToken(
			sg.NewTracedGRPCContext(ctx, span),
			token,
		),
		gReq,
	)
	if err != nil {
		span.AddLog(&sg.KV{"putTask error", err.Error()})
		return err
	}

	return nil
}

func (t *Topology) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	fmt.Println("{{TOPOLOGY}}", extras)
	return nil, &cPb.ListResp{Data: []*cPb.Data{}}
}
