.PHONY: build run clean test fmt vet deps api-gen

GOPROXY := $(shell go env GOPROXY)

# Build the application
build:
	go build -o bin/chat-rag main.go

# Build for Windows (with .exe suffix)
build-win:
	go build -o bin/chat-rag.exe main.go

# Run the application
run:
	go run main.go -f etc/chat-api.yaml

# Run with custom config
run-config:
	go run main.go -f $(CONFIG)

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf logs/

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Setup project (install tools and generate code)
setup: install-tools api-gen deps

# Development server with auto-reload (requires air)
dev:
	air

# Install air for development
install-air:
	go install github.com/cosmtrek/air@latest

# Docker build
docker-build:
	docker build -t chat-rag:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 chat-rag:latest

# Build, tag and push Docker image with version
docker-release:
	docker build -t chat-rag:$(VERSION) --build-arg IMAGE_VERSION=$(VERSION) \
	--build-arg GOPROXY=$(GOPROXY) \
	.
	docker tag chat-rag:$(VERSION) ${REGISTRY}/chat-rag:$(VERSION)
# 	docker push ${REGISTRY}/chat-rag:$(VERSION)

# Build the container image with the wasm plugin
build-release-aliyun:
	docker build -t chat-rag:$(VERSION) --build-arg IMAGE_VERSION=$(VERSION) \
		--build-arg GOPROXY=$(GOPROXY) \
		.
	docker tag chat-rag:$(VERSION) ${REGISTRY}/chat-rag:$(VERSION)
	docker push ${REGISTRY}/chat-rag:$(VERSION)

build-release:
	docker build -t chat-rag:$(VERSION) --build-arg IMAGE_VERSION=$(VERSION) \
	--build-arg GOPROXY=$(GOPROXY) \
	.
	docker tag chat-rag:$(VERSION) ${REGISTRY}/chat-rag:$(VERSION)
	cd ~/sangfor/upload-docker-images/images-zgsm/ && \
	rm -f * && \
	docker save -o chat-rag-$(VERSION).tar ${REGISTRY}/chat-rag:$(VERSION) && \
	git add -A && \
	git commit -m "feat: add chat-rag-$(VERSION).tar" && \
	git push origin main



# Create necessary directories
init-dirs:
	mkdir -p logs
	mkdir -p bin

# Full setup for new environment
bootstrap: install-tools init-dirs api-gen deps build
	@echo "Project setup complete!"
	@echo "Run 'make run' to start the server"

# Help
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application with default config"
	@echo "  run-config  - Run with custom config (CONFIG=path/to/config.yaml)"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  vet         - Vet code"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  api-gen     - Generate API code from .api file"
	@echo "  setup       - Install tools and generate code"
	@echo "  dev         - Run development server with auto-reload"
	@echo "  docker-release - Build, tag and push Docker image (VERSION=v1.0.0)"
	@echo "  bootstrap   - Full setup for new environment"
	@echo "  help        - Show this help message"