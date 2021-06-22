package main

import (
	"context"
	"fmt"
	pb "github.com/dbielecki97/grpc-go-course/greet/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"time"
)

type server struct {
}

func (s server) Greet(ctx context.Context, r *pb.GreetRequest) (*pb.GreetResponse, error) {
	fmt.Printf("Greet request was invoked with: %v\n", r)
	g := r.GetGreeting()
	return &pb.GreetResponse{Result: fmt.Sprintf("Hello %s %s", g.FirstName, g.LastName)}, nil
}

func (s server) GreetManyTimes(r *pb.GreetManyTimesRequest, stream pb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes was invoked")
	firstName := r.Greeting.FirstName
	for i := 0; i < 10; i++ {
		res := &pb.GreetManyTimesResponse{Result: fmt.Sprintf("Hello %v number %d", firstName, i)}
		stream.Send(res)
		time.Sleep(time.Second)
	}

	return nil
}

func (s server) LongGreet(stream pb.GreetService_LongGreetServer) error {
	result := ""
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&pb.LongGreetResponse{Result: result})
			}
			log.Fatalf("error while reading from client: %v", err)
		}
		result += "Hello " + req.Greeting.FirstName + "\n"
	}

	return nil
}

func (s server) GreetEveryone(stream pb.GreetService_GreetEveryoneServer) error {

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Fatalf("error while reading client stream: %v", err)
			return err
		}

		result := "Hello " + req.Greeting.FirstName + "! \n"
		err = stream.Send(&pb.GreetEveryoneResponse{Result: result})
		if err != nil {
			log.Fatalf("error while sending data to client: %v", err)
			return err
		}
	}
}

func (s server) GreetWithDeadline(ctx context.Context, r *pb.GreetWithDeadlineRequest) (*pb.GreetWithDeadlineResponse, error) {
	fmt.Printf("Greet request was invoked with: %v\n", r)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The client cancelled the request")
			return nil, status.Error(codes.DeadlineExceeded, "The client cancelled request")
		}
		time.Sleep(time.Second)
	}
	g := r.GetGreeting()
	return &pb.GreetWithDeadlineResponse{Result: fmt.Sprintf("Hello %s %s", g.FirstName, g.LastName)}, nil
}

func main() {
	fmt.Println("Hello world!")

	var opts []grpc.ServerOption
	tls := false
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
		}
		opts = append(opts, grpc.Creds(creds))
	}
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
