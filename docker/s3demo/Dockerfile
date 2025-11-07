FROM golang:1 AS build-env
RUN apt-get update
WORKDIR /s3demo
ADD s3demo/go.mod /s3demo/go.mod
ADD s3demo/go.sum /s3demo/go.sum
RUN go mod download
ADD ./s3demo/*.go /s3demo/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /s3demo
COPY --from=build-env /s3demo/server /s3demo/server
ADD ./s3demo/*.go.html /s3demo/
ADD ./static/osd /s3demo/osd
EXPOSE 8080
ENTRYPOINT ["/s3demo/server"]
