package vault

import (
	"context"
	"github.com/c12s/celestial/service"
	aPb "github.com/c12s/scheme/apollo"
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

func (s *SSecrets) getToken(ctx context.Context, user string) (error, string) {
	client := service.NewApolloClient(s.db.apolloAddress)
	resp, err := client.GetToken(ctx, &aPb.GetReq{user})
	if err != nil {
		return err, ""
	}
	return nil, resp.Token
}

func (s *SSecrets) List(ctx context.Context, key, user string) (error, map[string]string) {
	uerr, userId := s.getToken(ctx, user)
	if uerr != nil {
		return uerr, nil
	}

	s.db.init(userId)
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

func (s *SSecrets) Mutate(ctx context.Context, key, user string, req map[string]interface{}) (error, string) {
	uerr, userId := s.getToken(ctx, user)
	if uerr != nil {
		return uerr, ""
	}

	s.db.init(userId)
	defer s.db.revert()

	path := formatKey(key)
	_, err := s.db.client.Logical().Write(path, req)
	if err != nil {
		return err, ""
	}
	return nil, path
}
