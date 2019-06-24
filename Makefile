all: build test

build:
	go install github.com/parkr/github-utils/cmd/...

test:
	go test github.com/parkr/github-utils/...

docker-build:
	docker build -t parkr/github-utils .
