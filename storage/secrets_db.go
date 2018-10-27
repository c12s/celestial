package storage

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type SecretsDB interface {
	Secrets() Secrets
}

type Secrets interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}
