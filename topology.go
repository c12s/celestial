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

func marshall(node Node) []byte {
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

// Test if specified labels are present in node
func (self *Node) testLabels(kvs KVS) bool {
	have := true
	if len(self.Labels.Kvs) < len(kvs.Kvs) {
		return false
	}

	for k, v := range self.Labels.Kvs {
		if _, ok := dict[k]; !ok {
			have = false
			break
		}
	}

	return have
}

// If labesl are present, add new configs
func (self *Node) addConfigs(labels, data KVS) {
	if self.testLabels(labels.Kvs) {
		for k, v := range data.Kvs {
			self.Configs.Kvs[k] = v
		}
	}
}

func (self *Client) CreateConfig(regionid, clusterid, nodeid, jobid string, labels, configs KVS) {
	key := fmt.Sprintf("/topology/%s/%s/%s/", regionid, clusterid, nodeid)
	resp, _ := self.Kv.Get(self.Ctx, key)
	for _, item := range resp.Kvs {
		// Get node data from ETCD
		node := unmarshall(item.Value)

		// Update current configs with new ones
		node.addConfigs(labels, configs)

		// Save back to ETCD
		self.Kv.Put(self.Ctx, key, string(marshall(node)))

		//TODO: Notify some Task queue to push configs to the devices
	}
}

func (self *Client) CreateSecret(regionid, clusterid, nodeid string, labels, secrets KVS) {

}

func (self *Client) GetNodes(regionid, clusterid string) []Node {
	var nodes []Node
	nodesKey := fmt.Sprintf("/topology/%s/%s/") // /toplology/regionid/clusterid/

	gr, _ := self.Kv.Get(self.Ctx, nodesKey, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	for _, item := range gr.Kvs {
		node := unmarshall(item.Value)
		nodes = append(nodes, *node)
	}

	return nodes
}

func (self *Client) AddNode(clusterid, regionid, nodeid string, node Node) error {
	nodeKey := fmt.Sprintf("/toplology/%s/%s/%s") // /topology/regionid/clusterid/nodeid/
	nodeData := marshall(node)
	_, err := self.Kv.Put(self.Ctx, nodeKey, string(nodeData))

	return err
}

func (self *Client) Put() {
	k := "/topology/region/nodes"

	_, _ = self.Kv.Put(self.Ctx, k, "test1")
	_, _ = self.Kv.Put(self.Ctx, k, "test2")

	resp, _ := self.Kv.Get(self.Ctx, k)
	for _, item := range resp.Kvs {
		fmt.Println(string(item.Key), string(item.Value))
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

func (self *Client) Close() {
	self.Cli.Close()
}

func testStore() {
	for _, item := range test() {
		str := marshall(item)
		fmt.Println(string(str))
		fmt.Println("\n")

		n := unmarshall(str)
		fmt.Println(n)
		fmt.Println("\n")

	}
}

func main() {
	// client := NewClient(config.DefaultConfig())
	// defer client.Close()

	// client.GetNodes()
	// client.Put()

	testStore()
}
