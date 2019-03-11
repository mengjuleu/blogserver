package main

import (
	"fmt"
	"net/http"

	blogweb "github.com/blog/blogweb/server"
	"github.com/blog/healthpb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	grpcDefaultPort    = "50051"
	blogWebDefaultPort = "5000"
)

func main() {
	logger := logrus.New()
	logger.Info("Start blog web...")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial(fmt.Sprintf("localhost:%s", grpcDefaultPort), opts)
	if err != nil {
		logger.WithError(err).Errorf("Could not connect to port: %v", grpcDefaultPort)
	}
	defer cc.Close()

	c := healthpb.NewHealthClient(cc)

	s, err := blogweb.NewServer(
		blogweb.UseLogger(logrus.NewEntry(logger)),
		blogweb.UseHealthClient(c),
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create new server instance")
	}

	http.ListenAndServe(fmt.Sprintf(":%s", blogWebDefaultPort), s.Route())
}
