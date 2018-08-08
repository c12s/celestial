package etcd

import (
	"context"
	"github.com/c12s/celestial/model"
)

type Configs struct {
	db *DB
}

func (s *Configs) List(ctx context.Context, regionid, clusterid string, labels model.KVS) (error, []model.Node) {
	return nil, nil
}

func (s *Configs) Mutate(ctx context.Context, regionid, clusterid string, labels, data model.KVS) error {
	return nil
}
