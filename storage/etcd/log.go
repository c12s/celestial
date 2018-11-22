package etcd

import (
	"context"
	"github.com/c12s/celestial/helper"
	cPb "github.com/c12s/scheme/celestial"
	"github.com/golang/protobuf/proto"
)

func logMutate(ctx context.Context, req *cPb.MutateReq, db *DB) (string, error) {
	logKey := helper.TasksKey()
	data, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	_, err = db.Kv.Put(ctx, logKey, string(data))
	if err != nil {
		return "", err
	}

	return "Log done", nil
}
