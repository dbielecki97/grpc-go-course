package main

import (
	"context"
	"fmt"
	pb "github.com/dbielecki97/grpc-go-course/calculator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial: %v", err)
	}

	c := pb.NewCalculatorClient(cc)

	doSquare(c)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBiDiStreaming(c)
}

func doBiDiStreaming(c pb.CalculatorClient) {
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("error creating bidi stream: %v", err)
	}

	requests := []*pb.FindMaximumRequest{
		{Number: 10},
		{Number: 2},
		{Number: 3},
		{Number: 4},
		{Number: 4},
		{Number: 4},
		{Number: 12},
		{Number: 20},
		{Number: 19},
	}

	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			fmt.Printf("sending number: %v\n", req)
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("error receiving from server: %v", err)
			}

			fmt.Printf("received: %v\n", response.Maximum)
		}
		close(waitc)
	}()

	<-waitc
}

func doClientStreaming(c pb.CalculatorClient) {
	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Could not create stream for computing average: %v", err)
	}

	requests := []*pb.ComputeAverageRequest{
		{Number: 1},
		{Number: 2},
		{Number: 3},
		{Number: 4},
	}

	for _, req := range requests {
		log.Printf("Sending: %v", req)
		stream.Send(req)
		time.Sleep(time.Second)
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Could not receive average: %v", err)
	}

	fmt.Printf("Average: %f", result.Average)
}

func doServerStreaming(c pb.CalculatorClient) {
	stream, err := c.Decompose(context.Background(), &pb.DecomposeRequest{Number: 523788534})
	if err != nil {
		log.Fatalf("Error when calling Decompose: %v", err)
	}

	for {
		result, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error when reading stream: %v", err)
		}

		fmt.Printf("Received decomposition factor: %v\n", result.Number)
	}
}

func doUnary(c pb.CalculatorClient) {
	result, err := c.Sum(context.Background(), &pb.SumRequest{
		A: 10,
		B: 5,
	})
	if err != nil {
		log.Fatalf("Could not calculate sum: %v", err)
	}

	fmt.Println(result.Sum)
}

func doSquare(c pb.CalculatorClient) {
	result, err := c.SquareRoot(context.Background(), &pb.SquareRootRequest{
		Number: -5,
	})
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			if e.Code() == codes.InvalidArgument {
				log.Fatalf("We probably sent a negative number! %v", e.Message())
			} else {
				log.Fatalf("Code: %v, Message:%v", e.Code(), e.Message())
			}
		} else {
			log.Fatalf("Could not calculate sum: %v", err)
		}

	}

	fmt.Println(result.NumberRoot)
}
