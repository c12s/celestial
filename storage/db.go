package storage

import (
	"context"
	"github.com/c12s/celestial/model/storage"
)

type DB interface {
	Secrets() Secrets
	Configs() Configs
}

type Configs interface {
	List(ctx context.Context, regionid, clusterid string, labels storage.KVS) (error, storage.Node)
	Mutate(ctx context.Context, regionid, clusterid string, labels, data storage.KVS) error
}

type Secrets interface {
	List(ctx context.Context, regionid, clusterid string, labels storage.KVS) (error, storage.Node)
	Mutate(ctx context.Context, regionid, clusterid string, labels, data storage.KVS) error
}
