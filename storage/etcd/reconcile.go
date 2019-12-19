package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/helper"
	bPb "github.com/c12s/scheme/blackhole"
	fPb "github.com/c12s/scheme/flusher"
)

type Reconcile struct {
	db *DB
}

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
	r.db.s.Sub(func(msg *fPb.Update) {
		go func(data *fPb.Update) {
			//TODO: remove task from gravity or update
			//TODO: update node job status

			key := helper.ConstructKey(data.Node, data.Kind)
			kind := helper.ToUpper(data.Kind)
			value, ok := bPb.TaskKind_value[kind]
			if !ok {
				return
			}
			err := r.update(ctx, key, "Done", bPb.TaskKind(value))
			if err != nil {
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
