package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/c12s/celestial/config"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

/*

Topology:
  - RegionID: Name
  - ResourcesTotal:
      - CPU: 1234
      - Memory: 2345
      - Storate: 123445
  - ResourcesUsed:
      - CPU: 12354
      - Memory: 2345
      - Storate: 4356
  - Nodes:
      - Node1:
          - Labels:
              - Key1: Value1
              - Key2: Value2
          - Configs:
              - Key1: Value1
              - Key2: Value2
          - Secrets:
              - Secret1: Value1
              - Secret2: Value2
          - Jobs:
              - Process1:
                  - Configs:
                      - Key1: Value1
                      - Key2: Value2
                  - Secrets:
                      - Key1: Value1
                      - Key2: Value2
                  - Labels:
                      - Key1: Value1
                      - Key2: Value2

*/

type Client struct {
	Kv  clientv3.KV
	Ctx context.Context
	Cli clientv3.Client
}

type KVS struct {
	Kvs map[string]string `json:"kvs"`
}

type Job struct {
	Labels  KVS `json:"labels"`
	Configs KVS `json:"configs"`
	Secrets KVS `json:"secrets"`
}

type Node struct {
	Labels  KVS   `json:"labels"`
	Configs KVS   `json:"configs"`
	Secrets KVS   `json:"secrets"`
	Jobs    []Job `json:"jobs"`
}

func marshall(node *Node) []byte {
	bnode, err := json.Marshal(node)
	if err != nil {
		log.Fatal(err)
	}

	return bnode
}

func unmarshall(blob []byte) *Node {
	var node Node
	err := json.Unmarshal(blob, &node)
	if err != nil {
		log.Fatal(err)
	}

	return &node
}

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

func (self *Client) Close() {
	self.Cli.Close()
}

// Test if specified labels are present in node
func (self *Node) testLabels(labels KVS) bool {
	have := true
	if len(self.Labels.Kvs) < len(labels.Kvs) {
		return false
	}

	for k, _ := range labels.Kvs {
		if _, ok := self.Labels.Kvs[k]; !ok {
			have = false
			break
		}
	}

	return have
}

// If labesl are present, add new configs
func (self *Node) addConfigs(labels, data KVS) {
	if self.testLabels(labels) {
		for k, v := range data.Kvs {
			self.Configs.Kvs[k] = v
		}
	}
}

// Select Nodes that contains labels or key-value pares specified by user
func (self *Client) SelectNodes(clusterid, regionid string, selector KVS) []Nodes {
	var nodes = []Nodes{}

	key := fmt.Sprintf("/topology/%s/%s/", regionid, clusterid)
	for _, node := range self.GetClusterNodes(regionid, clusterid) {
		if node.testLabels(selector) {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// add new configuration to specific node
func (self *Client) CreateNodeConfig(regionid, clusterid, nodeid, jobid string, labels, configs KVS) {
	key := fmt.Sprintf("/topology/%s/%s/%s/", regionid, clusterid, nodeid)

	// Get node data from ETCD
	node := self.GetClusterNode(regionid, clusterid, nodeid)
	if node != nil {
		// Update current configs with new ones
		node.addConfigs(labels, configs)

		// Save back to ETCD
		self.Kv.Put(self.Ctx, key, string(marshall(node)))

		//TODO: Notify some Task queue to push configs to the devices
	} else {
		fmt.Println("Error")
	}
}

func (self *Client) CreateNodeSecret(regionid, clusterid, nodeid string, labels, secrets KVS) {

}

func (self *Client) GetClusterNodes(regionid, clusterid string) []Node {
	var nodes []Node
	nodesKey := fmt.Sprintf("/topology/%s/%s/", regionid, clusterid) // /toplology/regionid/clusterid/
	gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	for _, item := range gr.Kvs {
		node := unmarshall(item.Value)
		nodes = append(nodes, *node)
	}

	return nodes
}

func (self *Client) PrintClusterNodes(regionid, clusterid string) {
	nodesKey := fmt.Sprintf("/topology/%s/%s/", regionid, clusterid) // /toplology/regionid/clusterid/
	gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))

	for _, item := range gr.Kvs {
		fmt.Println(string(item.Key))
		fmt.Println(string(item.Value))
		fmt.Println("\n")
	}
}

func (self *Client) GetClusterNode(regionid, clusterid, nodeid string) *Node {
	key := fmt.Sprintf("/topology/%s/%s/%s/", regionid, clusterid, nodeid)
	resp, _ := self.Kv.Get(self.Ctx, key)

	for _, item := range resp.Kvs {
		node := unmarshall(item.Value)
		return node
	}

	return nil
}

func (self *Client) AddNode(regionid, clusterid, nodeid string, node *Node) int64 {
	nodeKey := fmt.Sprintf("/topology/%s/%s/%s/", regionid, clusterid, nodeid) // /topology/regionid/clusterid/nodeid/
	nodeData := marshall(node)
	pr, _ := self.Kv.Put(self.Ctx, nodeKey, string(nodeData))

	return pr.Header.Revision
}

func (self *Client) printEtcd() {
	gr, _ := self.Kv.Get(self.Ctx, "/topology/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	for _, item := range gr.Kvs {
		fmt.Println(string(item.Value))
	}
}

func test() []Node {
	l1 := make(map[string]string)
	l1["label1"] = "value1"
	l1["label2"] = "value2"
	l1["label3"] = "value3"
	l1["label4"] = "value4"

	l2 := make(map[string]string)
	l2["label1"] = "value1"
	l2["label2"] = "value2"

	c1 := make(map[string]string)
	c1["config1"] = "value1"
	c1["config2"] = "value2"
	c1["config3"] = "value3"
	c1["config4"] = "value4"

	c2 := make(map[string]string)
	c2["config1"] = "value1"
	c2["config2"] = "value2"

	s1 := make(map[string]string)
	s1["secret1"] = "value1"
	s1["secret2"] = "value2"

	s2 := make(map[string]string)
	s2["secret1"] = "value1"
	s2["secret2"] = "value2"
	s2["secret3"] = "value3"
	s2["secret4"] = "value4"

	j1 := Job{
		Labels:  KVS{Kvs: l1},
		Configs: KVS{Kvs: c1},
		Secrets: KVS{Kvs: s2},
	}

	j2 := Job{
		Labels:  KVS{Kvs: l2},
		Configs: KVS{Kvs: c1},
		Secrets: KVS{Kvs: s2},
	}

	node1 := Node{
		Labels:  KVS{Kvs: l1},
		Configs: KVS{Kvs: c2},
		Secrets: KVS{Kvs: s2},
		Jobs:    []Job{j1},
	}

	node2 := Node{
		Labels:  KVS{Kvs: l1},
		Configs: KVS{Kvs: c1},
		Secrets: KVS{Kvs: s2},
		Jobs:    []Job{j1, j2},
	}

	return []Node{node1, node2}
}

func testStore(c *Client) {
	for i, item := range test() {
		nodeid := fmt.Sprintf("node-%d", i)
		rev := c.AddNode("novisad", "grbavica", nodeid, &item)
		fmt.Println("Added", item, "Revision", rev)
	}
}

func testJson() {
	for _, n := range test() {
		s := marshall(&n)
		fmt.Println(string(s))
		fmt.Println("\n")
	}
}

func main() {
	client := NewClient(config.DefaultConfig())
	defer client.Close()

	client.PrintClusterNodes("novisad", "grbavica")
	// testStore(client)

	// testJson()

	//n := client.GetClusterNode("novisad", "grbavica", "node-1")
	//fmt.Println(n)
	//client.printEtcd()
}
