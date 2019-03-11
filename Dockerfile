FROM golang:1.11.5

RUN go version

ADD . /go/src/github.com/blog

WORKDIR /go/src/github.com/blog
RUN make install-grpc-server

EXPOSE 50051

ENTRYPOINT ["/go/bin/blog_server"]