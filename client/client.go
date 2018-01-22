package client

import (
	"context"
	"fmt"
	"github.com/c12s/celestial/config"
	"github.com/coreos/etcd/clientv3"
)

type Client struct {
	Kv  clientv3.KV
	Ctx context.Context
	Cli clientv3.Client
}

// Create a new celestial client connection to kev-value store
func NewClient(c *config.Config) *Client {
	fmt.Println(c.GetRequestTimeout(), c.GetDialTimeout(), c.GetEndpoints())
	ctx, _ := context.WithTimeout(context.Background(), c.GetRequestTimeout())
	cli, _ := clientv3.New(clientv3.Config{
		DialTimeout: c.GetDialTimeout(),
		Endpoints:   c.GetEndpoints(),
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
// Return Node chanel from witch you get data, at the end of process it close the chanel
func (self *Client) SelectNodes(clusterid, regionid string, selector KVS) <-chan Node {
	nodeChan := make(chan Node)
	go func() {
		for node := range self.GetClusterNodes(regionid, clusterid) {
			if node.TestLabels(selector) {
				nodeChan <- node
			}
		}
		close(nodeChan)
	}()

	return nodeChan
}

// Add new configuration to specific nodes
func (self *Client) MutateNodes(regionid, clusterid string, labels, data KVS, kind int) {
	key := GenerateKey(regionid, clusterid)
	// Get nodes data from ETCD that contains selector labels
	for node := range self.SelectNodes(regionid, clusterid, labels) {
		// Update current configs with new ones
		node.AddConfig(labels, data, kind)

		// Save back to ETCD
		self.Kv.Put(self.Ctx, key, string(node.Marshall()))

		//TODO: Notify some Task queue to push configs to the devices
	}
}

func (self *Client) nodesGenerator(regionid, clusterid string) <-chan Node {
	nodeChan := make(chan Node)
	go func() {
		for node := range self.GetClusterNodes(regionid, clusterid) {
			nodeChan <- node
		}
		close(nodeChan)
	}()

	return nodeChan
}

func (self *Client) jobesGenerator(regionid, clusterid string, selector, data KVS, kind int) <-chan Node {
	nodesChan := make(chan Node)
	go func() {
		for node := range self.GetClusterNodes(regionid, clusterid) {
			for job := range node.SelectJobs(selector) {
				// Update jobs configuration
				job.AddConfig(selector, data, kind)
			}
			nodesChan <- node
		}
	}()
	return nodesChan
}

// Add new configuration to specific jobs
func (self *Client) MutateJobs(regionid, clusterid string, selector, data KVS, kind int) {
	key := GenerateKey(regionid, clusterid)
	for node := range self.jobesGenerator(regionid, clusterid, selector, data, kind) {
		// Save back to ETCD
		self.Kv.Put(self.Ctx, key, string(node.Marshall()))

		//TODO: Notify some Task queue to push configs to the devices
	}
}

func (self *Client) GetClusterNodes(regionid, clusterid string) <-chan Node {
	nodeChan := make(chan Node)
	go func() {
		nodesKey := GenerateKey(regionid, clusterid) // /toplology/regionid/clusterid/
		gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		for _, item := range gr.Kvs {
			nodeChan <- *Unmarshall(item.Value)
		}
		close(nodeChan)
	}()
	return nodeChan
}

func (self *Client) PrintClusterNodes(regionid, clusterid string) <-chan string {
	nodesChan := make(chan string)
	go func() {
		nodesKey := GenerateKey(regionid, clusterid) // /toplology/regionid/clusterid/
		gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		for _, item := range gr.Kvs {
			nodesChan <- fmt.Sprintf("Key:%s\nData:\n%s\n", string(item.Key), string(item.Value))
		}
		close(nodesChan)
	}()
	return nodesChan
}

func (self *Client) AddNode(regionid, clusterid, nodeid string, node *Node) int64 {
	nodeKey := GenerateKey(regionid, clusterid, nodeid) // /topology/regionid/clusterid/nodeid/
	nodeData := node.Marshall()
	pr, _ := self.Kv.Put(self.Ctx, nodeKey, string(nodeData))

	return pr.Header.Revision
}
