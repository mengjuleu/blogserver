package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/blog/blogpb"

	"google.golang.org/grpc"
)

const grpcDefaultPort = "50051"

func main() {
	fmt.Println("Blog Client")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial(fmt.Sprintf("localhost:%s", grpcDefaultPort), opts)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	listBlog(c)
}

func createBlog(c blogpb.BlogServiceClient) {
	blog := &blogpb.Blog{
		AuthorId: "mleu",
		Title:    "My First Blog Post",
		Content:  "Content of the first blog",
	}
	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Blog has been created: %v", createBlogRes)
}

func readBlog(c blogpb.BlogServiceClient, blogID string) {
	readBlogReq := &blogpb.ReadBlogRequest{
		BlogId: blogID,
	}
	readBlogRes, rerr := c.ReadBlog(context.Background(), readBlogReq)
	if rerr != nil {
		fmt.Printf("Error happened while reading: %v", rerr)
	}
	fmt.Printf("Blog was read %v\n", readBlogRes)
}

func updateBlog(c blogpb.BlogServiceClient, blogID string) {
	newBlog := &blogpb.Blog{
		Id:       blogID,
		AuthorId: "mleu",
		Title:    "My First Blog Post (edited)",
		Content:  "Content of the first blog, with some awesome addition",
	}

	updateBlogRes, uerr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: newBlog,
	})
	if uerr != nil {
		fmt.Printf("Error happened while updating: %v", uerr)
	}
	fmt.Printf("Blog was updated %v\n", updateBlogRes)
}

func deleteBlog(c blogpb.BlogServiceClient, blogID string) {
	deleteBlogRes, derr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: blogID,
	})

	if derr != nil {
		fmt.Printf("Error happened while deleting: %v", derr)
	}
	fmt.Printf("Blog was deleted %v\n", deleteBlogRes)
}

func listBlog(c blogpb.BlogServiceClient) {
	stream, lerr := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})

	if lerr != nil {
		log.Fatalf("error while calling list blog: %v", lerr)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v ", err)
		}
		fmt.Println(res.GetBlog())
	}
}
