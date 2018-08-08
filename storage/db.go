package storage

import (
	"context"
	"github.com/c12s/celestial/client"
)

type DB interface {
	Secrets() Secrets
	Configs() Configs
}

type Configs interface {
	List(ctx context.Context, regionid, clusterid string, labels client.KVS) (error, client.Node)
	Mutate(ctx context.Context, regionid, clusterid string, labels, data client.KVS) error
}

type Secrets interface {
	List(ctx context.Context, regionid, clusterid string, labels klient.KVS) (error, client.Node)
	Mutate(ctx context.Context, regionid, clusterid string, labels, data klient.KVS) error
}
