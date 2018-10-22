package etcd

import (
	"github.com/c12s/celestial/storage"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type DB struct {
	Kv     clientv3.KV
	Client *clientv3.Client
}

func New(endpoints []string, timeout time.Duration) (*DB, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: timeout,
		Endpoints:   endpoints,
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

func (db *DB) Close() { db.Client.Close() }

func (db *DB) Secrets() storage.Secrets { return &Secrets{db} }

func (db *DB) Configs() storage.Configs { return &Configs{db} }

func (db *DB) Actions() storage.Actions { return &Actions{db} }

func (db *DB) Namespaces() storage.Namespaces { return &Namespaces{db} }
