package helper

import (
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

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func ProtoToKVS(req *pb.ListReq) model.KVS {
	m := make(map[string]string)

	for _, item := range req.Labels.Labels {
		m[item.Key] = item.Value
	}

	return model.KVS{
		Kvs: m,
	}

}

func ProtoToKVSMutate(req *pb.MutateReq) (model.KVS, model.KVS) {
	l := make(map[string]string)
	d := make(map[string]string)

	for _, item := range req.Labels.Labels {
		l[item.Key] = item.Value
	}

	for _, item := range req.Data.Data {
		d[item.Key] = item.Value
	}

	return model.KVS{
			Kvs: l,
		},
		model.KVS{
			Kvs: d,
		}

}
