FROM alpine:3.6

RUN adduser -D sqs-queue-operator
USER sqs-queue-operator

ADD tmp/_output/bin/sqs-queue-operator /usr/local/bin/sqs-queue-operator
