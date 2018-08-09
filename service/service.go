package service

import (
	"errors"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/model/config"
	pb "github.com/c12s/celestial/pb"
	"github.com/c12s/celestial/storage"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	db *storage.DB
}

func (s *Server) handleList(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	if req.Kind == pb.ReqKind_SECRETS {
		labels := helper.ProtoToKVS(req.Labels)
		err, resp := s.db.Secrets().List(ctx, req.RegionId, req.ClusterId, labels)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return resp, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		labels := helper.ProtoToKVS(req.Labels)
		err, resp := s.db.Configs().List(ctx, req.RegionId, req.ClusterId, req.Labels)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return resp, nil
	}
}

func (s *Server) handleMutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	labels := helper.ProtoToKVS(req.Labels)
	data := helper.ProtoToKVS(req.Data)

	if req.Kind == pb.ReqKind_SECRETS {
		err := s.db.Secrets().Mutate(ctx, req.RegionId, req.ClusterId, labels, data)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return &pb.MutateResp{
			error: "NONE",
		}, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		err := s.db.Configs().Mutate(ctx, req.RegionId, req.ClusterId, labels, data)
		if err != nil {
			return nil, err
		}

		return &pb.MutateResp{
			error: "NONE",
		}, nil
	}
}

func (s *Server) List(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	resp, err := s.handleList(ctx, req)
	return resp, err
}

func (s *Server) Mutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	resp, err := s.handleMutate(ctx, req)
	return resp, err
}

func Run(db *storage.DB, conf *config.ConnectionConfig) {
	if !conf.Standalone {
		lis, err := net.Listen("tcp", conf.Rpc)
		if err != nil {
			log.Fatalf("failed to initializa TCP listen: %v", err)
		}
		defer lis.Close()

		server := grpc.NewServer()
		roleServer := &Server{
			db: db,
		}
		pb.RegisterSecretsServer(server, roleServer)
		server.Serve(lis)
	} else {
		log.Fatal("RPC service runs in standalone mode!")
	}
}
