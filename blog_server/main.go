package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/blog/blog_server/server"
	"github.com/blog/blogpb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

const (
	mongoURI       = "mongodb://mleu:password123@ds113765.mlab.com:13765/blog"
	defaultRPCPort = "50051"
)

func main() {
	logger := logrus.New()

	logger.Info("Blog Service Started")

	client, cerr := configureMongoClient()
	if cerr != nil {
		logger.WithError(cerr).Fatal("Failed to configure Mongo Client")
	}

	connerr := client.Connect(context.TODO())
	if connerr != nil {
		logger.WithError(connerr).Fatal("Failed to connect to MongoDB")
	}
	defer func() {
		logger.Info("Disconnecting from MongoDB")
		err := client.Disconnect(context.TODO())
		if err != nil {
			logger.WithError(err).Fatal("Failed to disconnect to MongoDB")
		}
	}()

	collection := client.Database("blog").Collection("blog")

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", defaultRPCPort))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to listen to port: %v", defaultRPCPort)
	}

	defer lis.Close()

	blogServer, err := server.NewServer(
		server.UseCollection(collection),
		server.UseLogger(logrus.NewEntry(logger)),
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create blog server")
	}

	s := configureGrpcServer(blogServer)

	go func() {
		logger.Info("Sarting GRPC server")
		if err := s.Serve(lis); err != nil {
			logger.WithError(err).Fatal("Failed to serve TCP listener")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	logger.Println("Stopping the server")
	s.Stop()
	logger.Println("End of Program")
}

func configureMongoClient() (*mongo.Client, error) {
	return mongo.NewClient(options.Client().ApplyURI(mongoURI))
}

func configureGrpcServer(blogServer *server.Server) *grpc.Server {
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, blogServer)
	return s
}
