package etcd

import (
	"context"
	"errors"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/model"
)

type Configs struct {
	db *DB
}

func (s *Configs) List(ctx context.Context, regionid, clusterid string, labels model.KVS) (error, []model.Node) {
	return s.db.Select(ctx, regionId, clusterId, labels)
}

func (s *Configs) Mutate(ctx context.Context, regionid, clusterid string, labels, data model.KVS) error {
	key := helper.GenerateKey(regionId, clusterId)
	done, err := s.db.SelectAndUpdate(ctx, key, selector, data)
	if err != nil {
		return err
	}

	if !done {
		err := errors.New("Operation not done, try again!")
		if err != nil {
			return err
		}
	}

	return nil
}
