FROM golang:1.24.4 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

ENV GOPROXY https://goproxy.cn,direct

ENV GOSUMDB off

WORKDIR /build

COPY . .

RUN make build

FROM alpine:latest AS STANDARD

# 添加必要的运行时库和证书
# RUN apk --no-cache add ca-certificates tzdata

# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
# ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/bin/main /app/server

# 确保server文件有执行权限
RUN chmod +x /app/server

CMD ["./server"]