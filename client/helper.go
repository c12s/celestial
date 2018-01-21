package client

import (
	"encoding/json"
	"strings"
)

const SECRETS = 1
const CONFIGS = 2

func Unmarshall(blob []byte) *Node {
	var node Node
	err := json.Unmarshal(blob, &node)
	Check(err)

	return &node
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
