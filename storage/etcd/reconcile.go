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
	sg "github.com/c12s/stellar-go"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
)

type Reconcile struct {
	db *DB
}

const watchKey = "topology/regions/tasks"

func (r *Reconcile) update(ctx context.Context, key, status string, kind bPb.TaskKind) error {
	span, _ := sg.FromGRPCContext(ctx, "reconcile.start")
	defer span.Finish()
	fmt.Println(span)

	switch kind {
	case bPb.TaskKind_SECRETS:
		err := r.db.Secrets().StatusUpdate(sg.NewTracedGRPCContext(ctx, span), key, status)
		if err != nil {
			return err
		}
	case bPb.TaskKind_ACTIONS:
		err := r.db.Actions().StatusUpdate(sg.NewTracedGRPCContext(ctx, span), key, status)
		if err != nil {
			return err
		}
	case bPb.TaskKind_CONFIGS:
		err := r.db.Configs().StatusUpdate(sg.NewTracedGRPCContext(ctx, span), key, status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconcile) Start(ctx context.Context, address string) {
	go func(c context.Context) {
		span, _ := sg.FromGRPCContext(c, "reconcile.push.start")
		defer span.Finish()
		fmt.Println(span)

		client := service.NewGravityClient(address)
		watchChan := r.db.Client.Watch(c, watchKey, clientv3.WithPrefix())
		for {
			select {
			case result := <-watchChan:
				go func() {
					child := span.Child("reconcile.event")
					defer child.Finish()
					for _, ev := range result.Events {
						mReq := &cPb.MutateReq{}
						err := proto.Unmarshal(ev.Kv.Value, mReq)
						if err != nil {
							child.AddLog(&sg.KV{"unmarshall error", err.Error()})
							return
						}

						for _, key := range mReq.Index {
							r.update(sg.NewTracedGRPCContext(c, child), key, "In progress", mReq.Mutate.Kind)
						}

						req := &gPb.PutReq{
							Key:     string(ev.Kv.Key), //key to be deleted after push is done
							Task:    mReq,
							TaskKey: string(ev.Kv.Key),
						}

						_, err = client.PutTask(sg.NewTracedGRPCContext(c, child), req)
						if err != nil {
							child.AddLog(&sg.KV{"putTask error", err.Error()})
							continue
						}
					}
				}()

			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			}
		}
	}(ctx)

	r.db.s.Sub(func(msg *fPb.Update) {
		go func(data *fPb.Update) {
			span, _ := sg.FromGRPCContext(ctx, "reconcile.pull.start")
			defer span.Finish()
			fmt.Println(span)

			//TODO: remove task from gravity or update
			//TODO: update node job status

			key := helper.ConstructKey(data.Node, data.Kind)
			kind := helper.ToUpper(data.Kind)
			value, ok := bPb.TaskKind_value[kind]
			if !ok {
				span.AddLog(&sg.KV{"kind validate", "kind not valid"})
				return
			}
			err := r.update(sg.NewTracedGRPCContext(ctx, span), key, "Done", bPb.TaskKind(value))
			if err != nil {
				span.AddLog(&sg.KV{"update error", err.Error()})
				return
			}

			fmt.Println()
			fmt.Print("GET celestial: ")
			fmt.Println(msg)
			fmt.Println()
		}(msg)
	})
	fmt.Println("Started")
}
