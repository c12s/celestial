package test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	cli, _ := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{"0.0.0.0:2379"},
	})
	defer cli.Close()
	kv := clientv3.NewKV(cli)

	//GetSingleValueDemo(ctx, kv)
	PutValues(ctx, kv)
}

func GetSingleValueDemo(ctx context.Context, kv clientv3.KV) {
	fmt.Println("*** GetSingleValueDemo()")
	// Delete all keys
	kv.Delete(ctx, "key", clientv3.WithPrefix())

	// Insert a key value
	pr, _ := kv.Put(ctx, "key", "444")
	rev := pr.Header.Revision
	fmt.Println("Revision:", rev)

	gr, _ := kv.Get(ctx, "key")
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)

	// Modify the value of an existing key (create new revision)
	kv.Put(ctx, "key", "555")

	gr, _ = kv.Get(ctx, "key")
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)

	// Get the value of the previous revision
	gr, _ = kv.Get(ctx, "key", clientv3.WithRev(rev))
	fmt.Println("Value: ", string(gr.Kvs[0].Value), "Revision: ", gr.Header.Revision)
}

func PutValues(ctx context.Context, kv clientv3.KV) {
	nodesKey := "/topology/nodes/congis"

	pr, _ := kv.Put(ctx, nodesKey+"node1", "444")
	rev := pr.Header.Revision
	fmt.Println("Revision:", rev)

	pr, _ = kv.Put(ctx, nodesKey+"node2", "555")
	rev = pr.Header.Revision
	fmt.Println("Revision:", rev)

}
