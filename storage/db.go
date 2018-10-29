package storage

import (
	"context"
	cPb "github.com/c12s/scheme/celestial"
)

type DB interface {
	Secrets() Secrets
	Configs() Configs
	Actions() Actions
	Namespaces() Namespaces
}

type Configs interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}

type Actions interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}

type Namespaces interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}

type Secrets interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}
