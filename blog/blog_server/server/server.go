package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/dbielecki97/grpc-go-course/blog/blog_server/model"
	pb "github.com/dbielecki97/grpc-go-course/blog/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Server struct {
	Collection *mongo.Collection
}

func New(collection *mongo.Collection) *Server {
	return &Server{Collection: collection}
}

func (s *Server) CreateBlog(ctx context.Context, r *pb.CreateBlogRequest) (*pb.CreateBlogResponse, error) {
	blog := r.GetBlog()

	data := model.BlogItem{
		AuthorId: blog.AuthorId,
		Content:  blog.Content,
		Title:    blog.Title,
	}

	result, err := s.Collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot cast OID"))
	}

	res := &pb.CreateBlogResponse{Blog: &pb.Blog{
		Id:       id.Hex(),
		AuthorId: data.AuthorId,
		Title:    data.Title,
		Content:  data.Content,
	}}

	return res, nil
}

func (s *Server) ReadBlog(ctx context.Context, r *pb.ReadBlogRequest) (*pb.ReadBlogResponse, error) {
	oid, err := primitive.ObjectIDFromHex(r.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("could not create OID from blog_id string: %v", err))
	}

	result := s.Collection.FindOne(context.Background(), &bson.D{{"_id", oid}})
	mErr := result.Err()
	if mErr != nil {
		if errors.As(mErr, &mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot find blog with specified id"))
		}
	}

	var data model.BlogItem
	err = result.Decode(&data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot decode BlogItem: %v", err))
	}

	res := &pb.ReadBlogResponse{Blog: &pb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorId,
		Title:    data.Title,
		Content:  data.Content,
	}}
	return res, nil
}

func (s *Server) UpdateBlog(ctx context.Context, r *pb.UpdateBlogRequest) (*pb.UpdateBlogResponse, error) {
	blog := r.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		log.Printf("Could not create OID from string: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("id is not a hex format"))
	}

	item := &model.BlogItem{
		ID:       oid,
		AuthorId: blog.AuthorId,
		Content:  blog.Content,
		Title:    blog.Title,
	}
	result, err := s.Collection.ReplaceOne(context.Background(), &bson.D{{"_id", oid}}, item)
	if err != nil {
		log.Printf("could not replace BlogItem: %v", err)
		return nil, status.Errorf(codes.Internal, "unexpected database error")
	}

	if result.MatchedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "blog with specified id could not be found")
	}

	res := &pb.UpdateBlogResponse{Blog: &pb.Blog{
		Id:       blog.Id,
		AuthorId: blog.AuthorId,
		Title:    blog.Title,
		Content:  blog.Content,
	}}
	return res, nil
}

func (s *Server) DeleteBlog(ctx context.Context, r *pb.DeleteBlogRequest) (*pb.DeleteBlogResponse, error) {
	oid, err := primitive.ObjectIDFromHex(r.BlogId)
	if err != nil {
		log.Printf("Could not create OID from string: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("id is not a hex format"))
	}

	result, err := s.Collection.DeleteOne(context.Background(), &bson.D{{"_id", oid}})
	if err != nil {
		log.Printf("Could not delete BlogItem: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unexpected database error"))
	}

	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "blog with specified id could not be found")
	}

	return &pb.DeleteBlogResponse{BlogId: r.BlogId}, nil
}

func (s *Server) ListBlog(r *pb.ListBlogRequest, stream pb.BlogService_ListBlogServer) error {
	cur, err := s.Collection.Find(context.Background(), &bson.D{})
	if err != nil {
		log.Printf("Could not list BlogItem: %v", err)
		return status.Errorf(codes.Internal, fmt.Sprintf("unexpected database error"))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		time.Sleep(time.Second)
		var data model.BlogItem
		err := cur.Decode(&data)
		if err != nil {
			log.Printf("Could not decode BlogItem")
			return status.Errorf(codes.Internal, fmt.Sprintf("error while decoding data: %v", err))
		}

		err = stream.Send(&pb.ListBlogResponse{Blog: &pb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorId,
			Title:    data.Title,
			Content:  data.Content,
		}})
		if err != nil {
			log.Printf("Could not send BlogItem to stream")
			return status.Errorf(codes.Internal, fmt.Sprintf("error while decoding data: %v", err))
		}
	}
	return nil
}
