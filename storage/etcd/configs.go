package etcd

import (
	"context"
	"github.com/c12s/celestial/model/storage"
)

type Configs struct {
	db *DB
}

func (s *Configs) List(ctx context.Context, regionid, clusterid string, labels storage.KVS) (error, []storage.Node) {
	return nil, nil
}

func (s *Configd) Mutate(ctx context.Context, regionid, clusterid string, labels, data storage.KVS) error {
	return nil
}
