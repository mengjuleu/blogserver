package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/blog/blogpb"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/status"
)

// Server is blog server
type Server struct {
	collection *mongo.Collection
	logger     *logrus.Entry
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

// UseLogger assigned the logger attribute
func UseLogger(logger *logrus.Entry) func(*Server) error {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

// UseCollection assigned the collection attribute
func UseCollection(c *mongo.Collection) func(*Server) error {
	return func(s *Server) error {
		s.collection = c
		return nil
	}
}

// NewServer creates an instance of blog server
func NewServer(opts ...func(*Server) error) (*Server, error) {
	s := &Server{}

	for _, f := range opts {
		if err := f(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// CreateBlog creates a blog post and store in MongoDB
func (s *Server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	s.logger.Info("Create blog request")
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := s.collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprint("Cannot convert to OID"),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

// ReadBlog gets a blog post from MongoDB
func (s *Server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	s.logger.Info("Read blog request")

	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	data := &blogItem{}
	filter := bson.M{"_id": oid}

	res := s.collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog post with specified ID: %v", oid),
		)
	}

	return &blogpb.ReadBlogResponse{Blog: dataToBlogPb(data)}, nil
}

// UpdateBlog updates an existing blog post
func (s *Server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	s.logger.Info("Update blog request")
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}
	data := &blogItem{}
	filter := bson.M{"_id": oid}

	res := s.collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog post with specified ID: %v", oid),
		)
	}

	// udpate internal staruct
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, uerr := s.collection.ReplaceOne(context.Background(), filter, data)
	if uerr != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot update object in MongoDB: %v", uerr),
		)
	}

	return &blogpb.UpdateBlogResponse{Blog: dataToBlogPb(data)}, nil
}

// DeleteBlog deletes an existing blog post
func (s *Server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	s.logger.Info("delete blog request")

	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	filter := bson.M{"_id": oid}

	deleteRes, derr := s.collection.DeleteOne(context.Background(), filter)
	if derr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete blog post: %v", derr),
		)
	}

	if deleteRes.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog post in MongoDB: %v", derr),
		)
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}, nil
}

// ListBlog lists all of the posts in MongoDB
func (s *Server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	s.logger.Info("List blog post request")

	cursor, err := s.collection.Find(context.Background(), primitive.D{})
	if err != nil {
		return status.Error(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		data := &blogItem{}
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)
		}
		if serr := stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(data)}); serr != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while sending data to client: %v", serr),
			)
		}
	}

	if err := cursor.Err(); err != nil {
		return status.Error(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}
