package main

import (
	"context"
	pb "github.com/dbielecki97/grpc-go-course/blog/proto"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not dial: %v", err)
	}
	defer cc.Close()

	c := pb.NewBlogServiceClient(cc)

	blog, err := c.CreateBlog(context.Background(), &pb.CreateBlogRequest{Blog: &pb.Blog{
		AuthorId: "123",
		Title:    "Iron Man",
		Content:  "RDJ",
	}})
	if err != nil {
		log.Fatalf("Could not create a blog: %v", err)
	}
	log.Println("Created blog: ", blog.GetBlog())

	res, err := c.ReadBlog(context.Background(), &pb.ReadBlogRequest{BlogId: blog.Blog.Id})
	if err != nil {
		log.Fatalf("could not read blog: %v", err)
	}

	log.Println("Read blog: ", res.GetBlog())

	updateBlog, err := c.UpdateBlog(context.Background(), &pb.UpdateBlogRequest{Blog: &pb.Blog{
		Id:       res.Blog.Id,
		AuthorId: "123",
		Title:    "123",
		Content:  "123",
	}})
	if err != nil {
		log.Fatalf("Could not update blog: %v", err)
	}

	log.Println("Updated blog: ", updateBlog.Blog)

	deleteBlog, err := c.DeleteBlog(context.Background(), &pb.DeleteBlogRequest{BlogId: updateBlog.Blog.Id})
	if err != nil {
		log.Fatalf("Could not delete blog: %v", err)
	}

	log.Println("Deleted blog id", deleteBlog.BlogId)

	stream, err := c.ListBlog(context.Background(), &pb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Could not create stream of ListBlog: %v", err)
	}

	for {
		recv, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				stream.CloseSend()
				break
			} else {
				log.Fatalf("unexpected error reading stream: %v", err)
			}
		}

		log.Println("Received new blog: ", recv.Blog)
	}

}
