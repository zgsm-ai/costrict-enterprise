TAG ?= latest

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

.PHONY:test
test:
	go test ./internal/...

.PHONY:build
build:
	go mod tidy
	go build -ldflags="-s -w" -o ./bin/main ./cmd/main.go

.PHONY:docker
docker:
	docker build -t zgsm/codebase-querier:$(TAG) .
	docker push zgsm/codebase-querier:$(TAG)

.PHONY:lint
lint:
	golangci-lint run ./...