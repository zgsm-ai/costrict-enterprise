# build stage
FROM golang:alpine AS build
RUN apk update && apk add git
ADD . /src
WORKDIR /src
ENV CGO_ENABLED=0
RUN go build \
    -ldflags "-X github.com/zgsm-ai/cotun/share.BuildVersion=$(git describe --abbrev=0 --tags)" \
    -o /tmp/cotund server-main/*.go
# run stage
FROM scratch
LABEL maintainer="zgsm@sangfor.com.cn"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=build /tmp/cotund /app/cotund
ENTRYPOINT ["/app/cotund"]