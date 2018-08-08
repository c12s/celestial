package etcd

import (
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/storage"
	"github.com/coreos/etcd/clientv3"
)

type DB struct {
	Kv     *clientv3.Client
	Client *clientv3.Client
}

func New(c *config.ClientConfig) (*DB, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: c.GetDialTimeout(),
		Endpoints:   c.GetEndpoints(),
	})

	if err != nil {
		return nil, err
	}

	kv := clientv3.NewKV(cli)
	return &DB{
		Kv:     kv,
		Client: cli,
	}, nil
}

func (db *DB) Secrets() storage.Secrets { return &Secrets{db} }

func (db *DB) Configs() storage.Configs { return &Configs{db} }
