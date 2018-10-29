package vault

import (
	"github.com/c12s/celestial/storage"
	"github.com/hashicorp/vault/api"
	"time"
)

type DB struct {
	client *api.Client
}

func New(endpoints []string, timeout time.Duration) (*DB, error) {
	cli, err := api.NewClient(&api.Config{
		Address: endpoints[0],
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		client: cli,
	}, nil
}

func (db *DB) init(token string) {
	db.client.SetToken(token)
}

func (db *DB) revert() {
	db.client.SetToken("")
}

func (db *DB) Close() {}

func (db *DB) SSecrets() storage.SSecrets { return &SSecrets{db} }
