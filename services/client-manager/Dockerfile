FROM golang:1.24.0 AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go env -w CGO_ENABLED=0 && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct
RUN go mod download && go mod verify

COPY . .
ARG VERSION=v1.0.0

RUN go build -ldflags="-s -w -X 'main.SoftwareVer=$VERSION'" -o client-manager *.go
RUN chmod 755 client-manager

FROM alpine:3.21 AS runtime
ENV env prod
ENV TZ Asia/Shanghai
WORKDIR /
COPY --from=builder /app/client-manager /usr/local/bin
ENTRYPOINT ["/usr/local/bin/client-manager"]
