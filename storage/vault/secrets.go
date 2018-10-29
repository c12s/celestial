package vault

import (
	"context"
	"strings"
)

type SSecrets struct {
	db *DB
}

const (
	keyPrefix = "ckv/data"
)

//ckv/data/topology/regionid/clusterid/nodes/nodeid/secrets
func formatKey(path string) string {
	s := []string{keyPrefix, path}
	return strings.Join(s, "/")
}

func (s *SSecrets) List(ctx context.Context, key string) (error, map[string]string) {
	s.db.init("myroot")
	defer s.db.revert()

	retVal := map[string]string{}
	path := formatKey(key)
	secretValues, err := s.db.client.Logical().Read(path)
	if err != nil {
		return err, nil
	}

	if secretValues != nil {
		for propName, propValue := range secretValues.Data {
			retVal[propName] = propValue.(string)
		}
	}
	return nil, retVal
}

func (s *SSecrets) Mutate(ctx context.Context, key string, req map[string]interface{}) (error, string) {
	s.db.init("myroot")
	defer s.db.revert()

	path := formatKey(key)
	_, err := s.db.client.Logical().Write(path, req)
	if err != nil {
		return err, ""
	}
	return nil, path
}
