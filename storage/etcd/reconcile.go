package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/service"
	cPb "github.com/c12s/scheme/celestial"
	gPb "github.com/c12s/scheme/gravity"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
)

type Reconcile struct {
	db *DB
}

const watchKey = "topology/regions/tasks"

func (r *Reconcile) Start(ctx context.Context, address string) {
	fmt.Println("Reconcile started...")
	go func() {
		client := service.NewGravityClient(address)
		watchChan := r.db.Client.Watch(ctx, watchKey, clientv3.WithPrefix())
		for {
			select {
			case result := <-watchChan:
				for _, ev := range result.Events {
					mReq := &cPb.MutateReq{}
					err := proto.Unmarshal(ev.Kv.Value, mReq)
					if err != nil {
						fmt.Println(err)
						return
					}

					req := &gPb.PutReq{
						Key:  string(ev.Kv.Key),
						Task: mReq,
					}

					_, err = client.PutTask(ctx, req)
					if err != nil {
						fmt.Println(err)
					}

					fmt.Println(string(ev.Kv.Key))
					fmt.Println(req)

					//TODO: Maybe change status of the task from Waiting to In Progress or something like that...think about it
					// fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				}
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			}
		}
	}()
}
