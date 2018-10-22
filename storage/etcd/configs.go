package etcd

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type Configs struct {
	db *DB
}

func (c *Configs) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	return nil, nil
}

func (c *Configs) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	return nil, nil
}
