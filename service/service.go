package service

import (
	"fmt"
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

func (s *Server) List(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	resp, err := s.handleList(ctx, req)
	return resp, err
}

func (s *Server) Mutate(ctx context.Context, req *pb.MutateReq) (*pb.MutateResp, error) {
	resp, err := s.handleMutate(ctx, req)
	if err != nil {
		return &pb.MutateResp{
			Error: err.Error(),
		}, nil
	}

	return resp, err
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
