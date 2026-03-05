# Define build arguments
ARG IMAGE_NAME=chat-rag
ARG IMAGE_VERSION=latest

# Build stage
FROM golang:1.24.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

ARG GOPROXY
ENV GOPROXY=${GOPROXY}
# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-rag .

# Runtime stage
FROM alpine:latest AS runtime

# Redeclare build arguments for runtime stage
ARG IMAGE_NAME=chat-rag
ARG IMAGE_VERSION=latest

# Install timezone data and set China timezone
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Set working directory
WORKDIR /app

# Copy compiled binary from build stage
COPY --from=builder /app/chat-rag .

# Copy configuration files
COPY etc/chat-api.yaml ./etc/

# Expose application port
EXPOSE 8888

# Set entrypoint
ENTRYPOINT ["./chat-rag"]

# Set image labels
LABEL name="${IMAGE_NAME}"
LABEL version="${IMAGE_VERSION}"