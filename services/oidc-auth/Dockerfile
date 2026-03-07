FROM golang:1.23.9-alpine3.22 AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/main cmd/main.go

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata && \
    ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN addgroup -S appgroup -g 1001 && adduser -S appuser -G appgroup -u 1001

WORKDIR /app

RUN mkdir -p /app/logs /app/config && \
    chown -R appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /app/main .
COPY --chown=appuser:appgroup config/config.yaml ./config/

RUN chmod +x /app/main

USER 1001

EXPOSE 8080

ENTRYPOINT ["./main"]
CMD ["serve", "--config", "config/config.yaml"]