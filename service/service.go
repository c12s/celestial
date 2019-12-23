package service

import (
	"errors"
	"fmt"
	"github.com/c12s/celestial/helper"
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/storage"
	aPb "github.com/c12s/scheme/apollo"
	bPb "github.com/c12s/scheme/blackhole"
	cPb "github.com/c12s/scheme/celestial"
	sg "github.com/c12s/stellar-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type Server struct {
	db         storage.DB
	instrument map[string]string
	apollo     string
}

func (s *Server) List(ctx context.Context, req *cPb.ListReq) (*cPb.ListResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "celestial.list")
	defer span.Finish()
	fmt.Println(span)

	token, err := helper.ExtractToken(ctx)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return nil, err
	}

	client := NewApolloClient(s.apollo)
	resp, err := client.Auth(
		helper.AppendToken(
			sg.NewTracedGRPCContext(ctx, span),
			token,
		),
		&aPb.AuthOpt{
			Data: map[string]string{"intent": "auth"},
		},
	)
	if err != nil {
		span.AddLog(&sg.KV{"apollo resp error", err.Error()})
		return nil, err
	}

	if !resp.Value {
		span.AddLog(&sg.KV{"apollo.auth value", resp.Data["message"]})
		return nil, errors.New(resp.Data["message"])
	}

	switch req.Kind {
	case cPb.ReqKind_SECRETS:
		err, resp := s.db.Secrets().List(sg.NewTracedGRPCContext(ctx, span), req.Extras)
		if err != nil {
			span.AddLog(&sg.KV{"secrets list error", err.Error()})
			return nil, err
		}
		return resp, nil
	case cPb.ReqKind_ACTIONS:
		err, resp := s.db.Actions().List(sg.NewTracedGRPCContext(ctx, span), req.Extras)
		if err != nil {
			span.AddLog(&sg.KV{"action list error", err.Error()})
			return nil, err
		}
		return resp, nil
	case cPb.ReqKind_CONFIGS:
		err, resp := s.db.Configs().List(sg.NewTracedGRPCContext(ctx, span), req.Extras)
		if err != nil {
			span.AddLog(&sg.KV{"configs list error", err.Error()})
			return nil, err
		}
		return resp, nil
	case cPb.ReqKind_NAMESPACES:
		err, resp := s.db.Namespaces().List(sg.NewTracedGRPCContext(ctx, span), req.Extras)
		if err != nil {
			span.AddLog(&sg.KV{"namespaces list error", err.Error()})
			return nil, err
		}
		return resp, nil
	}
	return &cPb.ListResp{Error: "Not valid file type"}, nil
}

func (s *Server) Mutate(ctx context.Context, req *cPb.MutateReq) (*cPb.MutateResp, error) {
	span, _ := sg.FromGRPCContext(ctx, "celestial.mutate")
	defer span.Finish()
	fmt.Println(span)

	token, err := helper.ExtractToken(ctx)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return nil, err
	}

	client := NewApolloClient(s.apollo)
	resp, err := client.Auth(
		helper.AppendToken(
			sg.NewTracedGRPCContext(ctx, span),
			token,
		),
		&aPb.AuthOpt{
			Data: map[string]string{"intent": "auth"},
		},
	)
	if err != nil {
		span.AddLog(&sg.KV{"apollo resp error", err.Error()})
		return nil, err
	}

	if !resp.Value {
		span.AddLog(&sg.KV{"apollo.auth value", resp.Data["message"]})
		return nil, errors.New(resp.Data["message"])
	}

	switch req.Mutate.Kind {
	case bPb.TaskKind_SECRETS:
		err, resp := s.db.Secrets().Mutate(sg.NewTracedGRPCContext(ctx, span), req)
		if err != nil {
			span.AddLog(&sg.KV{"secrets mutate error", err.Error()})
			return nil, err
		}
		return resp, nil
	case bPb.TaskKind_ACTIONS:
		err, resp := s.db.Actions().Mutate(sg.NewTracedGRPCContext(ctx, span), req)
		if err != nil {
			span.AddLog(&sg.KV{"actions mutate error", err.Error()})
			return nil, err
		}
		return resp, nil
	case bPb.TaskKind_CONFIGS:
		err, resp := s.db.Configs().Mutate(sg.NewTracedGRPCContext(ctx, span), req)
		if err != nil {
			span.AddLog(&sg.KV{"configs mutate error", err.Error()})
			return nil, err
		}
		return resp, nil
	case bPb.TaskKind_NAMESPACES:
		err, resp := s.db.Namespaces().Mutate(sg.NewTracedGRPCContext(ctx, span), req)
		if err != nil {
			span.AddLog(&sg.KV{"namespaces mutate error", err.Error()})
			return nil, err
		}
		return resp, nil
	}
	return &cPb.MutateResp{Error: "Not valid file type"}, nil
}

func Run(db storage.DB, conf *config.Config) {
	lis, err := net.Listen("tcp", conf.Address)
	if err != nil {
		log.Fatalf("failed to initializa TCP listen: %v", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	celestialServer := &Server{
		db:         db,
		instrument: conf.InstrumentConf,
		apollo:     conf.Apollo,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n, err := sg.NewCollector(celestialServer.instrument["address"], celestialServer.instrument["stopic"])
	if err != nil {
		fmt.Println(err)
		return
	}
	c, err := sg.InitCollector(celestialServer.instrument["location"], n)
	if err != nil {
		fmt.Println(err)
		return
	}
	go c.Start(ctx, 15*time.Second)
	db.Reconcile().Start(ctx, conf.Gravity)

	fmt.Println("Celestial RPC Started")
	cPb.RegisterCelestialServiceServer(server, celestialServer)
	server.Serve(lis)
}
