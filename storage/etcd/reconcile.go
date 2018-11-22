package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
)

type Reconcile struct {
	db *DB
}

const watchKey = "topology/regions/tasks"

func (r *Reconcile) Start(ctx context.Context) {
	fmt.Println("Reconcile started...")
	go func() {
		watchChan := r.db.Client.Watch(ctx, watchKey, clientv3.WithPrefix())
		for {
			select {
			case result := <-watchChan:
				for _, ev := range result.Events {
					fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				}
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			}
		}
	}()
}
