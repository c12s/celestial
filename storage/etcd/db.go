package etcd

import (
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/storage"
	sync "github.com/c12s/celestial/storage/sync"
	"github.com/c12s/celestial/storage/sync/nats"
	"github.com/c12s/celestial/storage/vault"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type DB struct {
	Kv      clientv3.KV
	Client  *clientv3.Client
	sdb     storage.SecretsDB
	s       sync.Syncer
	Gravity string
	Apollo  string
}

func New(conf *config.Config, timeout time.Duration) (*DB, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: timeout,
		Endpoints:   conf.Endpoints,
	})

	if err != nil {
		return nil, err
	}

	//Load secrets database
	sdb, err := vault.New(conf.SEndpoints, timeout, conf.Apollo)
	if err != nil {
		return nil, err
	}

	ns, err := nats.NewNatsSync(conf.Syncer, conf.STopic)
	if err != nil {
		return nil, err
	}

	return &DB{
		Kv:      clientv3.NewKV(cli),
		Client:  cli,
		sdb:     sdb,
		s:       ns,
		Gravity: conf.Gravity,
		Apollo:  conf.Apollo,
	}, nil
}

func (db *DB) Close() { db.Client.Close() }

func (db *DB) Secrets() storage.Secrets { return &Secrets{db} }

func (db *DB) Configs() storage.Configs { return &Configs{db} }

func (db *DB) Actions() storage.Actions { return &Actions{db} }

func (db *DB) Namespaces() storage.Namespaces { return &Namespaces{db} }

func (db *DB) Reconcile() storage.Reconcile { return &Reconcile{db} }
