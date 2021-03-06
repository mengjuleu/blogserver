FROM golang:1.11.5-alpine

RUN apk add --no-cache git make curl
RUN go version

RUN curl https://glide.sh/get | sh

ADD . /go/src/github.com/blog

WORKDIR /go/src/github.com/blog
RUN make install-grpc-server

EXPOSE 50051

ENTRYPOINT ["/go/bin/blog_server"]