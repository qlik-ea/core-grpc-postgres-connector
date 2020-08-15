FROM golang:1.15 AS builder
RUN go get -u github.com/golang/dep/cmd/dep
ADD . /go/src/github.com/qlik-oss/core-grpc-postgres-connector
WORKDIR /go/src/github.com/qlik-oss/core-grpc-postgres-connector
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/qlik-oss/core-grpc-postgres-connector/main .
CMD ["./main"]
EXPOSE 50051
