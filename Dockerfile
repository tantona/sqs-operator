FROM golang:latest

WORKDIR /go/src/github.com/tantona/sqs-operator
COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o sqs-operator ./cmd/sqs-operator/*.go

FROM alpine:3.6
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=0 /go/src/github.com/tantona/sqs-operator /usr/local/bin/

CMD "sqs-operator"