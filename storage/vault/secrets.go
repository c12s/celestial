package vault

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type Secrets struct {
	db *DB
}

func (s *Secrets) List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp) {
	return nil, nil
}

func (s *Secrets) Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp) {
	return nil, nil
}
