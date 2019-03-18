all: test lint

.gotlint:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	touch $@

.gotglide:
	go get github.com/Masterminds/glide
	touch $@

.gotdeps: .gotglide glide.lock
	glide install
	touch $@

test: .gotdeps
	go test -race -v ./...

install-grpc-server: .gotdeps
	go install github.com/blog/blog_server

install-web-server: .gotdeps
	go install github.com/blog/blogweb

lint: .gotlint
	gometalinter --fast --vendor \
	--enable gofmt \
	--disable gotype \
	--disable gocyclo \
	--exclude="file permissions" --exclude="Errors unhandled" \
	./...
