package vault

import (
	"github.com/c12s/celestial/storage"
	"github.com/hashicorp/vault/api"
	"time"
)

type DB struct {
	Client *api.Client
}

func New(endpoints []string, timeout time.Duration) (*DB, error) {
	cli, err := api.NewClient(&api.Config{
		Address: endpoints[0],
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		Client: cli,
	}, nil
}

func (db *DB) Close() {}

func (db *DB) Secrets() storage.Secrets { return &Secrets{db} }
