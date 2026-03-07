TAG ?= latest

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

.PHONY:mock
mock:
	mockgen -source=internal/store/vector/vector_store.go -destination=internal/store/vector/mocks/vector_store_mock.go --package=mocks

.PHONY:test
test:
	go test ./internal/...

.PHONY:build
build:
	go mod tidy
	go build -ldflags="-s -w" -o ./bin/main ./cmd/main.go

.PHONY:docker
docker:
	docker build -t zgsm/codebase-embedder:$(TAG) .
	docker push zgsm/codebase-embedder:$(TAG)

.PHONY:lint
lint:
	golangci-lint run ./...