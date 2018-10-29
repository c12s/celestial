package storage

import (
	"context"
)

type SecretsDB interface {
	SSecrets() SSecrets
}

type SSecrets interface {
	List(ctx context.Context, path string) (error, map[string]string)
	Mutate(ctx context.Context, key string, req map[string]interface{}) (error, string)
}
