package helper

import (
	"fmt"
	"strings"
)

const (
	SECRETS = 1
	CONFIGS = 2
)

func GenerateKey(data ...string) string {
	key := "/topology/%s/"
	return fmt.Sprintf(key, strings.Join(data, "/"))
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
