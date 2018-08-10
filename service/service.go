package service

import (
	"fmt"
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
	db storage.DB
}

func (s *Server) handleList(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	if req.Kind == pb.ReqKind_SECRETS {
		labels := helper.ProtoToKVS(req)
		err, _ := s.db.Secrets().List(ctx, req.RegionId, req.ClusterId, labels)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return &pb.ListResp{
			Error: "NONE",
			Data:  []*pb.KV{},
		}, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		labels := helper.ProtoToKVS(req)
		err, _ := s.db.Configs().List(ctx, req.RegionId, req.ClusterId, labels)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return &pb.ListResp{
			Error: "NONE",
			Data:  []*pb.KV{},
		}, nil
	}

	return &pb.ListResp{
		Error: "Not valid file type",
		Data:  []*pb.KV{},
	}, nil
}

func (s *Server) handleMutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	labels, data := helper.ProtoToKVSMutate(req)
	if req.Kind == pb.ReqKind_SECRETS {
		err := s.db.Secrets().Mutate(ctx, req.RegionId, req.ClusterId, labels, data)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		return &pb.MutateResp{
			Error: "NONE",
		}, nil
	}

	if req.Kind == pb.ReqKind_CONFIGS {
		err := s.db.Configs().Mutate(ctx, req.RegionId, req.ClusterId, labels, data)
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

func (s *Server) List(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	resp, err := s.handleList(ctx, req)
	return resp, err
}

func (s *Server) Mutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	// resp, err := s.handleMutate(ctx, req)
	// if err != nil {
	// 	return &pb.MutateResp{
	// 		error: err.Error(),
	// 	}, nil
	// }

	// return resp, err

	return &pb.MutateResp{
		Error: "OK",
	}, nil
}

func Run(db storage.DB, conf *config.ConnectionConfig) {
	if !conf.Standalone {
		lis, err := net.Listen("tcp", conf.Rpc.Address)
		if err != nil {
			log.Fatalf("failed to initializa TCP listen: %v", err)
		}
		defer lis.Close()

		server := grpc.NewServer()
		celestialServer := &Server{
			db: db,
		}

		fmt.Println("Celestial RPC Started")
		pb.RegisterCelestialServiceServer(server, celestialServer)
		server.Serve(lis)
	} else {
		log.Fatal("RPC service runs in standalone mode!")
	}
}
