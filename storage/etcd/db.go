package etcd

import (
	"context"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/model"
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/storage"
	"github.com/coreos/etcd/clientv3"
	"log"
)

type DB struct {
	Kv     clientv3.KV
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

func (db *DB) Select(ctx context.Context, regionId, clusterId string, selector model.KVS) (error, []model.Node) {
	//First selct only nodes that are inside cluster!
	gr, err := db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		log.Fatal(err)
		return err, nil
	}

	nodes := []model.Node{}
	for _, item := range gr.Kvs {
		node, err := helper.NodeUnmarshall(item.Value)
		if err != nil {
			return err, nil
		}

		nodes = append(nodes, node)
	}

	return nodes, err
}

func (db *DB) SelectAndUpdate(ctx context.Context, key string, selector, data model.KVS) (bool, error) {
	//First selct only nodes that are inside cluster!
	gr, err := db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	//Than start the transaction... because all nodes must be updated!
	txn, err = db.Kvs.Txn(ctx)
	if err != nil {
		return false, err
	}

	for _, item := range gr.Kvs {
		node, err := helper.NodeUnmarshall(item.Value)
		if err != nil {
			return false, nil
		}

		//Test if specific node contains all of given labels
		// It that is the case than updatem oterwise not!
		if node.TestLabels(selector) {
			nodeKey := string(item.Key)
			data, err := helper.NodeMarshall(node)
			if err != nil {
				return false, err
			}

			//Update node
			data.AddConfig(data, helper.CONFIGS)

			//Convert to string and start transaction
			nodeValue := string(data)
			txn = txn.Then(
				OpPut(nodeKey, nodeValue),
			)
		}
	}

	putresp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	return putresp.Succeeded, nil
}

func (db *DB) Close() { db.Client.Close() }

func (db *DB) Secrets() storage.Secrets { return &Secrets{db} }

func (db *DB) Configs() storage.Configs { return &Configs{db} }
