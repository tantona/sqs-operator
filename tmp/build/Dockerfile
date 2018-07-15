FROM alpine:3.6

RUN adduser -D sqs-operator
USER sqs-operator

ADD tmp/_output/bin/sqs-operator /usr/local/bin/sqs-operator
