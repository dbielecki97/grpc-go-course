package main

import (
	"context"
	"fmt"
	pb "github.com/dbielecki97/grpc-go-course/greet/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello I'a a client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := pb.NewGreetServiceClient(cc)

	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBiDirectionalStreaming(c)
	doUnaryWithDeadline(c, time.Second*5)
	doUnaryWithDeadline(c, time.Second)
}

func doBiDirectionalStreaming(c pb.GreetServiceClient) {
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error creating bidi stream: %v", err)
	}

	requests := []*pb.GreetEveryoneRequest{
		{Greeting: &pb.Greeting{FirstName: "Dawid", LastName: "Bielecki"}},
		{Greeting: &pb.Greeting{FirstName: "Mateusz", LastName: "Bielecki"}},
		{Greeting: &pb.Greeting{FirstName: "Krzysztof", LastName: "Bielecki"}},
	}

	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			fmt.Printf("sending message: %v\n", req)
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
				close(waitc)
				log.Fatalf("error receiving from server: %v", err)
			}

			fmt.Printf("received: %v", response.Result)
		}
		close(waitc)
	}()

	<-waitc
}

func doClientStreaming(c pb.GreetServiceClient) {
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Could not create LongGreet client")
	}

	requests := []*pb.LongGreetRequest{
		{Greeting: &pb.Greeting{FirstName: "Dawid", LastName: "Bielecki"}},
		{Greeting: &pb.Greeting{FirstName: "Dawid", LastName: "Bielecki"}},
		{Greeting: &pb.Greeting{FirstName: "Dawid", LastName: "Bielecki"}},
	}

	for _, req := range requests {
		stream.Send(req)
		time.Sleep(time.Second)
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Could not reveive LongGreeting: %v", err)
	}

	fmt.Println(result.Result)
}

func doServerStreaming(c pb.GreetServiceClient) {
	fmt.Println("Starting do to a server streaming rpc..")

	stream, err := c.GreetManyTimes(context.Background(), &pb.GreetManyTimesRequest{Greeting: &pb.Greeting{
		FirstName: "Dawid",
		LastName:  "Bielecki",
	}})
	if err != nil {
		log.Fatalf("Error while calling server streaming GreetManyTimes rpc: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("error while reading stream: %v", err)
		}

		log.Printf("Response from GreetManyTimes: %v", msg.Result)
	}
}

func doUnary(c pb.GreetServiceClient) {
	req := &pb.GreetRequest{Greeting: &pb.Greeting{
		FirstName: "Dawid",
		LastName:  "Bielecki",
	}}
	response, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to greet: %v", err)
	}
	fmt.Printf(response.Result)
}

func doUnaryWithDeadline(c pb.GreetServiceClient, seconds time.Duration) {
	req := &pb.GreetWithDeadlineRequest{Greeting: &pb.Greeting{
		FirstName: "Dawid",
		LastName:  "Bielecki",
	}}
	ctx, cancel := context.WithTimeout(context.Background(), seconds)
	defer cancel()
	response, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline exceeded")
			} else {
				fmt.Printf("Unexpected error: %v\n", statusErr)
			}
		} else {
			log.Fatalf("Failed to greet: %v", err)
		}
		return
	}
	fmt.Printf("%v\n", response.Result)
}
