package helper

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/c12s/celestial/model"
	pb "github.com/c12s/celestial/pb"
	"strings"
)

const (
	SECRETS = 1
	CONFIGS = 2
)

// Marshall node informations into byte array
func NodeMarshall(self model.Node) ([]byte, error) {
	bnode, err := json.Marshal(self)
	if err != nil {
		return nil, err
	}

	return bnode, nil
}

func NodeUnmarshall(blob []byte) (*model.Node, error) {
	var node model.Node
	err := json.Unmarshal(blob, &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func GenerateKey(data ...string) string {
	key := "/topology/%s/"
	return fmt.Sprintf(key, strings.Join(data, "/"))
}

func GenerateKeys(regions, clusters []string) []string {
	keys := []string{}
	for _, region := range regions {
		for _, cluster := range clusters {
			key := GenerateKey(region, cluster)
			keys = append(keys, key)
		}
	}

	return keys
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func NodeToProto(resp []model.Node) *pb.ListResp {
	data := []*pb.NodeData{}

	for _, item := range resp {
		kvs := []*pb.KV{}
		for k, v := range item.Configs.Kvs {
			kv := &pb.KV{
				Key:   k,
				Value: v,
			}
			kvs = append(kvs, kv)
		}

		node := &pb.NodeData{
			NodeId: hashNode(item),
			Data:   kvs,
		}
		data = append(data, node)
	}

	return &pb.ListResp{
		Error: "NONE",
		Data:  data,
	}
}

func listReq(req *pb.ListReq) *model.KVS {
	m := make(map[string]string)
	for _, item := range req.Labels {
		m[item.Key] = item.Value
	}

	return &model.KVS{
		Kvs: m,
	}
}

func mutateReq(req *pb.MutateReq) (*model.KVS, *model.KVS, []string, []string) {
	l := make(map[string]string)
	d := make(map[string]string)

	for _, item := range req.Content.Labels {
		l[item.Key] = item.Value
	}

	for _, item := range req.Content.Data {
		d[item.Key] = item.Value
	}

	return &model.KVS{Kvs: l}, &model.KVS{Kvs: d},
		req.RegionIds, req.ClusterIds

}

func ProtoToKVS(req interface{}, data ...interface{}) {
	switch castReq := req.(type) {
	case pb.MutateReq:
		l, d, r, c := mutateReq(&castReq)
		data[0] = l
		data[1] = d
		data[2] = r
		data[3] = c
	case pb.ListReq:
		data[0] = listReq(&castReq)
	default:
		fmt.Println("NOt valid type")
	}
}

func hashNode(node model.Node) string {
	arrBytes := []byte{}
	jsonBytes, _ := json.Marshal(node)
	arrBytes = append(arrBytes, jsonBytes...)

	return fmt.Sprintf("%x", md5.Sum(arrBytes))
}
