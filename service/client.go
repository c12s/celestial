package service

import (
	aPb "github.com/c12s/scheme/apollo"
	gPb "github.com/c12s/scheme/gravity"
	rPb "github.com/c12s/scheme/magnetar"
	mPb "github.com/c12s/scheme/meridian"
	"google.golang.org/grpc"
	"log"
)

func NewApolloClient(address string) aPb.ApolloServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to apollo service: %v", err)
	}
	return aPb.NewApolloServiceClient(conn)
}

func NewGravityClient(address string) gPb.GravityServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to gravity service: %v", err)
	}
	return gPb.NewGravityServiceClient(conn)
}

func NewMeridianClient(address string) mPb.MeridianServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to meridian service: %v", err)
	}
	return mPb.NewMeridianServiceClient(conn)
}

func NewMagnetarClient(address string) rPb.MagnetarServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connection to meridian service: %v", err)
	}
	return rPb.NewMagnetarServiceClient(conn)
}
