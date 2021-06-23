package main

import (
	"github.com/dbielecki97/grpc-go-course/blog/blog_server/db"
	"github.com/dbielecki97/grpc-go-course/blog/blog_server/server"
	pb "github.com/dbielecki97/grpc-go-course/blog/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Blog service started")

	c, closeDb := db.New()
	defer closeDb()
	collection := c.Database("mydb").Collection("blog")
	srv := server.New(collection)

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Could not lister: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterBlogServiceServer(s, srv)

	go func() {
		log.Println("Starting server...")
		if err = s.Serve(lis); err != nil {
			log.Fatalf("Could not server: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	log.Println("Stopping the server...")
	s.Stop()
	log.Println("Stopping listener...")
	lis.Close()
	log.Println("Stopping program...")
}
