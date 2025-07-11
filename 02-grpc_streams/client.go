package main

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"

	deliverypb "examples/gen/go/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor),
	)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	client := deliverypb.NewDeliveryServiceClient(conn)

	// Unary call
	doUnary(client)

	// Server streaming
	doServerStreaming(client)

	// Client streaming
	doClientStreaming(client)

	// Bidirectional streaming
	doBidirectionalStreaming(client)
}

func doUnary(client deliverypb.DeliveryServiceClient) {
	log.Println("Unary RPC: GetPackageInfo")
	resp, err := client.GetPackageInfo(context.Background(), &deliverypb.GetPackageInfoRequest{Id: 42})
	if err != nil {
		log.Fatalf("error calling GetPackageInfo: %v", err)
	}
	log.Printf("Response: %+v\n", resp.Package)
}

func doServerStreaming(client deliverypb.DeliveryServiceClient) {
	log.Println("Server Streaming RPC: GetNewPackages")
	stream, err := client.GetNewPackages(context.Background(), &deliverypb.GetPackageInfoRequest{Id: 100})
	if err != nil {
		log.Fatalf("error starting stream: %v", err)
	}

	for {
		pkg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error receiving from stream: %v", err)
		}
		log.Printf("Received: %+v", pkg)
	}
}

func doClientStreaming(client deliverypb.DeliveryServiceClient) {
	log.Println("Client Streaming RPC: AddPackages")
	stream, err := client.AddPackages(context.Background())
	if err != nil {
		log.Fatalf("error starting client stream: %v", err)
	}

	for i := 0; i < 3; i++ {
		pkg := &deliverypb.Package{
			Id:          int32(i + 1),
			Description: "Client-streamed package",
			WeightKg:    float32(i+1) * 1.1,
		}
		log.Printf("Sending package: %+v", pkg)
		if err := stream.Send(pkg); err != nil {
			log.Fatalf("error sending package: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error receiving response: %v", err)
	}
	log.Printf("AddPackages response: %d packages added", resp.AddedCount)
}

func doBidirectionalStreaming(client deliverypb.DeliveryServiceClient) {
	log.Println("Bidirectional Streaming RPC: GetPackages")
	stream, err := client.GetPackages(context.Background())
	if err != nil {
		log.Fatalf("error starting bidi stream: %v", err)
	}

	done := make(chan struct{})

	// Receive responses
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Println("Server closed the stream.")
				break
			}
			if err != nil {
				log.Fatalf("error receiving: %v", err)
			}
			log.Printf("Received response: %+v", resp.Package)
		}
		close(done)
	}()

	// Send requests
	for i := 1; i <= 3; i++ {
		req := &deliverypb.GetPackageInfoRequest{Id: int32(i * 10)}
		log.Printf("Sending request: %v", req)
		if err := stream.Send(req); err != nil {
			log.Fatalf("error sending: %v", err)
		}
		time.Sleep(time.Second)
	}
	stream.CloseSend()
	<-done
}

func unaryClientInterceptor(
	ctx context.Context, method string,
	req, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	start := time.Now()
	log.Printf("--> %s | Request: %+v", method, req)

	err := invoker(ctx, method, req, reply, cc, opts...)

	log.Printf("<-- %s | Duration: %s | Error: %v | Response: %+v",
		method, time.Since(start), err, reply)
	return err
}
