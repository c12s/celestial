package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/service"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	gPb "github.com/c12s/scheme/gravity"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
)

type Reconcile struct {
	db *DB
}

const watchKey = "topology/regions/tasks"

func (r *Reconcile) update(ctx context.Context, key, status string, kind bPb.TaskKind) error {
	switch kind {
	case bPb.TaskKind_SECRETS:
		err := r.db.Secrets().StatusUpdate(ctx, key, status)
		if err != nil {
			return err
		}
	case bPb.TaskKind_ACTIONS:
		err := r.db.Actions().StatusUpdate(ctx, key, status)
		if err != nil {
			return err
		}
	case bPb.TaskKind_CONFIGS:
		err := r.db.Configs().StatusUpdate(ctx, key, status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconcile) Start(ctx context.Context, address string) {
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

					for _, key := range mReq.Index {
						r.update(ctx, key, "In progress", mReq.Mutate.Kind)
					}

					req := &gPb.PutReq{
						Key:     string(ev.Kv.Key), //key to be deleted after push is done
						Task:    mReq,
						TaskKey: string(ev.Kv.Key),
					}

					_, err = client.PutTask(ctx, req)
					if err != nil {
						fmt.Println(err)
						continue
					}
				}
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			}
		}
	}()
}
