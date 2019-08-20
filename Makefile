.PHONY: all
all: build test

.PHONY: build
build:
	go install github.com/parkr/github-utils/cmd/...

.PHONY: test
test:
	go test github.com/parkr/github-utils/...

.PHONY: docker-build
docker-build:
	docker build -t parkr/github-utils .
