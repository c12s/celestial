package etcd

import (
	"context"
	"github.com/c12s/celestial/model"
)

type Secrets struct {
	db *DB
}

func (s *Secrets) List(ctx context.Context, regionid, clusterid string, labels model.KVS) (error, []model.Node) {
	return nil, nil
}

func (s *Secrets) Mutate(ctx context.Context, regionids, clusterids []string, labels, data model.KVS) error {
	return nil
}
