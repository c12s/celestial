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

func (db *DB) Select(ctx context.Context, key string, selector model.KVS) (error, []model.Node) {
	//First selct only nodes that are inside cluster!
	gr, err := db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
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

		nodes = append(nodes, *node)
	}

	return err, nodes
}

func (db *DB) SelectAndUpdate(ctx context.Context, keys []string, selector, data model.KVS) (bool, error) {
	for _, key := range keys {
		_, err := db.Update(ctx, key, selector, data)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (db *DB) Update(ctx context.Context, key string, selector, data model.KVS) (bool, error) {
	//First selct only nodes that are inside cluster!
	gr, err := db.Kv.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	//Than start the transaction... because all nodes must be updated!
	txn := db.Kv.Txn(ctx)
	for _, item := range gr.Kvs {
		node, err := helper.NodeUnmarshall(item.Value)
		if err != nil {
			return false, nil
		}

		//Test if specific node contains all of given labels
		// It that is the case than updatem oterwise not!
		if node.TestLabels(selector) {
			//Update node
			node.AddConfig(data, helper.CONFIGS)

			//Do conversion
			data, err := helper.NodeMarshall(*node)
			if err != nil {
				return false, err
			}

			//start transaction
			nodeValue := string(data)
			nodeKey := string(item.Value)
			txn = txn.Then(
				clientv3.OpPut(nodeKey, nodeValue),
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
