package storage

import (
	"context"
	"github.com/c12s/celestial/model"
)

type DB interface {
	Secrets() Secrets
	Configs() Configs
}

type Configs interface {
	List(ctx context.Context, regionid, clusterid string, labels model.KVS) (error, []model.Node)
	Mutate(ctx context.Context, regionids, clusterids []string, labels, data model.KVS) error
}

type Secrets interface {
	List(ctx context.Context, regionid, clusterid string, labels model.KVS) (error, []model.Node)
	Mutate(ctx context.Context, regionids, clusterids []string, labels, data model.KVS) error
}
