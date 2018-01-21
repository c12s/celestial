package client

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/config"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type Client struct {
	Kv  clientv3.KV
	Ctx context.Context
	Cli clientv3.Client
}

// Create a new celestial client connection to kev-value store
func NewClient(c *config.Config) *Client {
	dialTimeout := time.Duration(c.RequestTimeout) * time.Second
	requestTimeout := time.Duration(c.DialTimeout) * time.Second

	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	cli, _ := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   c.Endpoints,
	})
	kv := clientv3.NewKV(cli)

	return &Client{
		Kv:  kv,
		Ctx: ctx,
		Cli: *cli,
	}
}

// Close existing celestial client connection to key-value store
func (self *Client) Close() {
	self.Cli.Close()
}

// Select Nodes that contains labels or key-value pairs specified by user
func (self *Client) SelectNodes(clusterid, regionid string, selector KVS) []Node {
	nodes := []Node{}

	for _, node := range self.GetClusterNodes(regionid, clusterid) {
		if node.TestLabels(selector) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Add new configuration to specific nodes
func (self *Client) MutateNodes(regionid, clusterid string, labels, data KVS, kind int) {
	key := GenerateKey(regionid, clusterid)

	// Get nodes data from ETCD that contains selector labels
	for _, node := range self.SelectNodes(regionid, clusterid, labels) {
		// Update current configs with new ones
		node.AddConfig(labels, data, kind)

		// Save back to ETCD
		self.Kv.Put(self.Ctx, key, string(node.Marshall()))

		//TODO: Notify some Task queue to push configs to the devices
	}
}

// Add new configuration to specific jobs
func (self *Client) MutateJobs(regionid, clusterid string, selector, data KVS, kind int) {
	key := GenerateKey(regionid, clusterid)

	// Get nodes from specific cluster fomr ETCD
	for _, node := range self.GetClusterNodes(regionid, clusterid) {
		// Select jobs that contains selector labels
		for _, job := range node.SelectJobs(selector) {
			// Update jobs configuration
			job.AddConfig(selector, data, kind)

			// Save back to ETCD
			self.Kv.Put(self.Ctx, key, string(node.Marshall()))

			//TODO: Notify some Task queue to push configs to the devices
		}
	}
}

func (self *Client) GetClusterNodes(regionid, clusterid string) []Node {
	nodesKey := GenerateKey(regionid, clusterid) // /toplology/regionid/clusterid/
	nodes := []Node{}

	gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	for _, item := range gr.Kvs {
		node := Unmarshall(item.Value)
		nodes = append(nodes, *node)
	}

	return nodes
}

func (self *Client) PrintClusterNodes(regionid, clusterid string) {
	nodesKey := GenerateKey(regionid, clusterid) // /toplology/regionid/clusterid/

	gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	for _, item := range gr.Kvs {
		fmt.Println(string(item.Key))
		fmt.Println(string(item.Value))
		fmt.Println("\n")
	}
}

func (self *Client) GetClusterNode(regionid, clusterid, nodeid string) *Node {
	key := GenerateKey(regionid, clusterid, nodeid)

	resp, _ := self.Kv.Get(self.Ctx, key)

	for _, item := range resp.Kvs {
		node := Unmarshall(item.Value)
		return node
	}

	return nil
}

func (self *Client) AddNode(regionid, clusterid, nodeid string, node *Node) int64 {
	nodeKey := GenerateKey(regionid, clusterid, nodeid) // /topology/regionid/clusterid/nodeid/
	nodeData := node.Marshall()
	pr, _ := self.Kv.Put(self.Ctx, nodeKey, string(nodeData))

	return pr.Header.Revision
}
