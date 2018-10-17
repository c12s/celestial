package service

import (
	"fmt"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/model"
	pb "github.com/c12s/celestial/pb"
	"golang.org/x/net/context"
	"log"
)

func (s *Server) handleList(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	var labels model.KVS
	helper.ProtoToKVS(req, &labels)
	if req.Kind == pb.ReqKind_SECRETS {
		err, _ := s.db.Secrets().List(ctx, req.RegionId, req.ClusterId, labels)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
			return nil, err
		}

		return &pb.ListResp{
			Error: "NONE",
			Data:  []*pb.NodeData{},
		}, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		err, resp := s.db.Configs().List(ctx, req.RegionId, req.ClusterId, labels)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		data := helper.NodeToProto(resp)
		return data, nil
	}

	return &pb.ListResp{
		Error: "Not valid file type",
		Data:  []*pb.NodeData{},
	}, nil
}

func (s *Server) handleMutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	var labels, data model.KVS
	var regions, clusters []string

	helper.ProtoToKVS(req, &labels, &data, regions, clusters)
	if req.Kind == pb.ReqKind_SECRETS {
		err := s.db.Secrets().Mutate(ctx, regions, clusters, labels, data)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return &pb.MutateResp{
			Error: "NONE",
		}, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		err := s.db.Configs().Mutate(ctx, regions, clusters, labels, data)
		if err != nil {
			return nil, err
		}

		return &pb.MutateResp{
			Error: "NONE",
		}, nil
	}

	return &pb.MutateResp{
		Error: "Not valid file type",
	}, nil
}
