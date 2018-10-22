package etcd

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type Actions struct {
	db *DB
}

func (a *Actions) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	return nil, nil
}

func (a *Actions) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	return nil, nil
}
