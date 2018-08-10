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
	key := helper.GenerateKey(regionid, clusterid)
	return s.db.Select(ctx, key, labels)
}

func (s *Configs) Mutate(ctx context.Context, regionid, clusterid string, labels, data model.KVS) error {
	key := helper.GenerateKey(regionid, clusterid)
	done, err := s.db.SelectAndUpdate(ctx, key, labels, data)
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
