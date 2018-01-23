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
	Cli *clientv3.Client
}

// Create a new celestial client connection to kev-value store
func NewClient(c *config.ClientConfig) *Client {
	ctx, _ := context.WithTimeout(context.Background(), c.GetRequestTimeout())
	cli, _ := clientv3.New(clientv3.Config{
		DialTimeout: c.GetDialTimeout(),
		Endpoints:   c.GetEndpoints(),
	})
	kv := clientv3.NewKV(cli)

	return &Client{
		Kv:  kv,
		Ctx: ctx,
		Cli: cli,
	}
}

// Close existing celestial client connection to key-value store
func (self *Client) Close() {
	self.Cli.Close()
}

// Select Nodes that contains labels or key-value pairs specified by user
// Return Node chanel from witch you get data, at the end of process it close the chanel
func (self *Client) selectNodes(clusterid, regionid string, selector KVS) <-chan Node {
	nodeChan := make(chan Node)
	go func() {
		for node := range self.GetClusterNodes(regionid, clusterid) {
			if node.testLabels(selector) {
				nodeChan <- node
			}
		}
		close(nodeChan)
	}()
	return nodeChan
}

// Add new configuration to specific nodes
func (self *Client) MutateNodes(regionid, clusterid string, labels, data KVS, kind int) {
	key := generateKey(regionid, clusterid)
	// Get nodes data from ETCD that contains selector labels
	for node := range self.selectNodes(regionid, clusterid, labels) {
		// Update current configs with new ones
		node.addConfig(labels, data, kind)

		// Save back to ETCD
		_, err := self.Kv.Put(self.Ctx, key, string(node.marshall()))
		if err != nil {
			fmt.Println(err)
		}

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
			for job := range node.selectJobs(selector) {
				// Update jobs configuration
				job.addConfig(selector, data, kind)
			}
			nodesChan <- node
		}
		close(nodesChan)
	}()
	return nodesChan
}

// Add new configuration to specific jobs
func (self *Client) MutateJobs(regionid, clusterid string, selector, data KVS, kind int) {
	key := generateKey(regionid, clusterid)
	for node := range self.jobesGenerator(regionid, clusterid, selector, data, kind) {
		// Save back to ETCD
		_, err := self.Kv.Put(self.Ctx, key, string(node.marshall()))
		if err != nil {
			fmt.Println(err)
		}

		//TODO: Notify some Task queue to push configs to the devices
	}
}

func (self *Client) GetClusterNodes(regionid, clusterid string) <-chan Node {
	nodeChan := make(chan Node)
	go func() {
		nodesKey := generateKey(regionid, clusterid) // /toplology/regionid/clusterid/
		gr, err := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		if err != nil {
			fmt.Println(err)
		}

		for _, item := range gr.Kvs {
			nodeChan <- unmarshall(item.Value)
		}
		close(nodeChan)
	}()
	return nodeChan
}

func (self *Client) PrintClusterNodes(regionid, clusterid string) <-chan string {
	nodesChan := make(chan string)
	go func() {
		nodesKey := generateKey(regionid, clusterid) // /toplology/regionid/clusterid/
		gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		for _, item := range gr.Kvs {
			nodesChan <- fmt.Sprintf("Key:%s\nData:\n%s\n", string(item.Key), string(item.Value))
		}
		close(nodesChan)
	}()
	return nodesChan
}

func (self *Client) AddNode(regionid, clusterid, nodeid string, node *Node) int64 {
	nodeKey := generateKey(regionid, clusterid, nodeid) // /topology/regionid/clusterid/nodeid/
	nodeData := node.marshall()
	pr, err := self.Kv.Put(self.Ctx, nodeKey, string(nodeData))
	if err != nil {
		fmt.Println(err)
	}

	return pr.Header.Revision
}

// Return all configurations for specific node in some cluster at some region
// Returned values is key-value par where key is name of config and value is config value
func (self *Client) NodeConfigs(regionid, clusterid, nodeid string) (KVS, error) {
	nodeKey := generateKey(regionid, clusterid, nodeid)
	resp, err := self.Kv.Get(self.Ctx, nodeKey)
	if err != nil {
		return KVS{Kvs: nil}, err
	}

	for _, ev := range resp.Kvs {
		n := unmarshall(ev.Value)
		return n.Configs, nil
	}

	return KVS{Kvs: nil}, nil
}

// Return all secrets for specific node in some cluster at some region
// Returned values is key-value par where key is name of secret and value is secret value
func (self *Client) NodeSecrets(regionid, clusterid, nodeid string) (KVS, error) {
	nodeKey := generateKey(regionid, clusterid, nodeid)
	resp, err := self.Kv.Get(self.Ctx, nodeKey)
	if err != nil {
		return KVS{Kvs: nil}, err
	}

	for _, ev := range resp.Kvs {
		n := unmarshall(ev.Value)
		return n.Secrets, nil
	}

	return KVS{Kvs: nil}, nil
}

func (self *Client) GetJobConfigs(regionid, clusterid, nodeid, jobid string) (KVS, error) {
	nodeKey := generateKey(regionid, clusterid, nodeid)
	_, err := self.Kv.Get(self.Ctx, nodeKey)
	if err != nil {
		return KVS{Kvs: nil}, err
	}

	// TODO: Think what to do about jobs configs and secrets!
	// Should everything go in different key, or store jobs in map

	return KVS{Kvs: nil}, err
}
