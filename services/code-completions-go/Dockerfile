FROM golang:1.24.0 AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go env -w CGO_ENABLED=0 && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct
RUN go mod download && go mod verify

COPY . .

ARG VERSION=v1.7.76
RUN go build -ldflags="-s -w -X 'main.SoftwareVer=$VERSION'" -o code-completion *.go
RUN chmod 755 code-completion

# FROM frolvlad/alpine-glibc:alpine-3.21_glibc-2.41 AS runtime
FROM alpine:3.21 AS runtime

ENV env prod
ENV TZ Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/code-completion /app/code-completion
COPY --from=builder /app/bin/deepseek-tokenizer /app/bin/deepseek-tokenizer

ENTRYPOINT ["/app/code-completion"]
