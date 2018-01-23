package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	SECRETS = 1
	CONFIGS = 2
)

func unmarshall(blob []byte) Node {
	var node Node
	err := json.Unmarshal(blob, &node)
	Check(err)

	return node
}

func generateKey(data ...string) string {
	key := "/topology/%s/"
	return fmt.Sprintf(key, strings.Join(data, "/"))
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
