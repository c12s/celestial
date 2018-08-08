package etcd

import (
	"context"
	"github.com/c12s/celestial/model/storage"
	"strings"
	"time"
)

type Secrets struct {
	db *DB
}

func (s *Secrets) List(ctx context.Context, regionid, clusterid string, labels storage.KVS) (error, []storage.Node) {
	return nil, nil
}

func (s *Secrets) Mutate(ctx context.Context, regionid, clusterid string, labels, data storage.KVS) error {
	return nil
}
