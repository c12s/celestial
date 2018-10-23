package service

import (
	"fmt"
	"github.com/c12s/celestial/storage"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	db storage.DB
}

func (s *Server) List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error) {
	switch req.Kind {
	case cPb.ReqKind_SECRETS:
	case cPb.ReqKind_ACTIONS:
	case cPb.ReqKind_CONFIGS:
	case cPb.ReqKind_NAMESPACES:
		err, resp := s.db.Namespaces().List(ctx, req.Extras)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return &cPb.ListResp{Error: "Not valid file type"}, nil
}

func (s *Server) Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error) {
	switch req.Mutate.Kind {
	case bPb.TaskKind_SECRETS:
	case bPb.TaskKind_ACTIONS:
	case bPb.TaskKind_CONFIGS:
	case bPb.TaskKind_NAMESPACES:
		err, resp := s.db.Namespaces().Mutate(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return &cPb.MutateResp{Error: "Not valid file type"}, nil
}

func Run(db storage.DB, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to initializa TCP listen: %v", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	celestialServer := &Server{
		db: db,
	}

	fmt.Println("Celestial RPC Started")
	cPb.RegisterCelestialServiceServer(server, celestialServer)
	server.Serve(lis)
}
