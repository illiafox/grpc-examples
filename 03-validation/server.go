package main

import (
	"buf.build/go/protovalidate"
	"context"
	deliverypb "examples/gen/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"io"
	"log"
	"net"
	"time"
)

type DeliveryService struct {
	deliverypb.UnimplementedDeliveryServiceServer
}

func (s *DeliveryService) GetPackageInfo(ctx context.Context, req *deliverypb.GetPackageInfoRequest) (
	*deliverypb.GetPackageInfoResponse, error,
) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	log.Printf("Received request for ID: %d", req.Id)
	return &deliverypb.GetPackageInfoResponse{
		Package: &deliverypb.Package{
			Id:                        req.Id,
			Description:               "Sample package",
			WeightKg:                  2.5,
			FragileNote:               nil,
			EstimatedDeliveryDuration: durationpb.New(time.Hour * 9999), // will not be delivered...
		},
	}, nil
}

func (s *DeliveryService) GetNewPackages(req *deliverypb.GetPackageInfoRequest, stream deliverypb.DeliveryService_GetNewPackagesServer) error {
	for i := 1; i <= 3; i++ {
		pkg := &deliverypb.Package{
			Id:                        req.Id + int32(i),
			Description:               "Streamed package",
			WeightKg:                  float32(i),
			EstimatedDeliveryDuration: durationpb.New(time.Minute * time.Duration(i*10)),
		}
		if err := stream.Send(pkg); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (s *DeliveryService) AddPackages(stream deliverypb.DeliveryService_AddPackagesServer) error {
	var count int32
	for {
		pkg, err := stream.Recv()
		if err == io.EOF {
			log.Println("All packages received.")
			return stream.SendAndClose(&deliverypb.AddPackageResponse{
				AddedCount: count,
			})
		}
		if err != nil {
			return err
		}
		log.Printf("Received package: %+v", pkg)
		count++
	}
}

func (s *DeliveryService) GetPackages(stream deliverypb.DeliveryService_GetPackagesServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Bidirectional stream finished.")
			return nil
		}
		if err != nil {
			return err
		}

		resp := &deliverypb.GetPackageInfoResponse{
			Package: &deliverypb.Package{
				Id:                        req.Id,
				Description:               "Bi-directional response",
				WeightKg:                  1.5,
				EstimatedDeliveryDuration: durationpb.New(time.Minute * 15),
			},
		}
		if err = stream.Send(resp); err != nil {
			return err
		}
	}
}

func unaryServerInterceptor(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	log.Printf("--> %s | Payload: %+v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	log.Printf("<-- %s | Duration: %s | Error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func streamServerInterceptor(
	srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	start := time.Now()
	log.Printf("stream --> %s | IsClientStream: %t | IsServerStream: %t",
		info.FullMethod, info.IsClientStream, info.IsServerStream)
	err := handler(srv, ss)
	log.Printf("stream <-- %s | Duration: %s | Error: %v", info.FullMethod, time.Since(start), err)
	return err
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryServerInterceptor),
		grpc.StreamInterceptor(streamServerInterceptor),
	}

	s := grpc.NewServer(opts...)
	deliverypb.RegisterDeliveryServiceServer(s, &DeliveryService{})

	log.Println("gRPC server listening on :50051")
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
