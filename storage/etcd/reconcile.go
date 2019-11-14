package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/service"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	fPb "github.com/c12s/scheme/flusher"
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

	r.db.s.Sub(func(msg *fPb.Update) {
		go func(data *fPb.Update) {
			//TODO: remove task from gravity or update
			//TODO: update node job status

			key := helper.ConstructKey(data.Node, data.Kind)
			kind := helper.ToUpper(data.Kind)
			value, ok := bPb.TaskKind_value[kind]
			if !ok {
				fmt.Println("Not valid") //TODO: Add to some log/trace
				return
			}
			err := r.update(ctx, key, "Done", bPb.TaskKind(value))
			if err != nil {
				return //TODO: Add to some log/trace
			}

			fmt.Println()
			fmt.Print("GET celestial: ")
			fmt.Println(msg)
			fmt.Println()
		}(msg)
	})
	fmt.Println("Started")
}
