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
	Reconcile() Reconcile
}

type Configs interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
	StatusUpdate(ctx context.Context, key, newStatus string) error
}

type Actions interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
	StatusUpdate(ctx context.Context, key, newStatus string) error
}

type Namespaces interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
}

type Secrets interface {
	List(ctx context.Context, extras map[string]string) (error, *cPb.ListResp)
	Mutate(ctx context.Context, req *cPb.MutateReq) (error, *cPb.MutateResp)
	StatusUpdate(ctx context.Context, key, newStatus string) error
}

type Reconcile interface {
	Start(ctx context.Context, gravity string)
}
