package etcd

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/client"
	"github.com/coreos/etcd/clientv3"
	"strings"
	"time"
)

type Secrets struct {
	KV     *clientv3.Client
	Client *clientv3.Client
}

func (s *Secrets) List(ctx context.Context, regionid, clusterid string, labels client.KVS) (error, []client.Node) {
	ss := []string{regionid, clusterid}
	key := strings.Join(ss, "/")
	rez, err := s.KV.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	if err != nil {
		return err, nil
	}

	for k, v := range rez.Kvs {
	}

}

func (s *Secrets) Mutate(ctx context.Context, regionid, clusterid string, labels, data client.KVS) error {

}
