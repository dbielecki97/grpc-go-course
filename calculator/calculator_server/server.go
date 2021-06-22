package main

import (
	"context"
	"fmt"
	pb "github.com/dbielecki97/grpc-go-course/calculator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
)

type server struct {
}

func (s server) Sum(ctx context.Context, r *pb.SumRequest) (*pb.SumResponse, error) {
	fmt.Printf("Greet request was invoked with: %v\n", r)
	return &pb.SumResponse{Sum: r.A + r.B}, nil
}

func (s server) Decompose(r *pb.DecomposeRequest, stream pb.Calculator_DecomposeServer) error {
	k := 2
	N := int(r.Number)

	for N > 1 {
		if N%k == 0 {
			stream.Send(&pb.DecomposeResponse{Number: int32(k)})
			N /= k
		} else {
			k += 1
		}
	}
	return nil
}

func (s server) ComputeAverage(stream pb.Calculator_ComputeAverageServer) error {
	var average float32
	var count int
	for {
		result, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&pb.ComputeAverageResponse{Average: average / float32(count)})
			}

			log.Fatalf("Error receiving from client: %v", err)
		}

		log.Printf("Received: %v", result)
		average += float32(result.Number)
		count++
	}

}

func (s server) FindMaximum(stream pb.Calculator_FindMaximumServer) error {
	max := int32(0)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Fatalf("could not read client stream")
			return err
		}

		number := req.Number
		if number > max {
			max = number
			err = stream.Send(&pb.FindMaximumResponse{Maximum: max})
		}

		if err != nil {
			log.Fatalf("could not send response to client: %v", err)
			return err
		}
	}
}

func (s server) SquareRoot(ctx context.Context, r *pb.SquareRootRequest) (*pb.SquareRootResponse, error) {
	if r.Number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("number cannot be less that a zero, %d < 0", r.Number))
	}
	root := math.Sqrt(float64(r.Number))
	return &pb.SquareRootResponse{NumberRoot: root}, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Could not lister: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterCalculatorServer(s, &server{})

	if err = s.Serve(lis); err != nil {
		log.Fatalf("Could not server: %v", err)
	}
}
