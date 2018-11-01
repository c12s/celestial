package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	// "strings"
	// "time"
)

var keys = [...]string{"topology/regions/secrets", "topology/regions/configs", "topology/regions/actions"}

type Reconcile struct {
	db *DB
}

func (r *Reconcile) startWorker(ctx context.Context, key string) {
	go func() {
		watchChan := r.db.Client.Watch(ctx, key, clientv3.WithPrefix())
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

func (r *Reconcile) Start(ctx context.Context) {
	for _, key := range keys {
		r.startWorker(ctx, key)
	}
}
