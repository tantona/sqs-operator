# SQS Operator

This operator allows you to manage AWS SQS queues with CustomResourceDefinitions in Kubernetes.


## Getting Started

```
dep ensure -v
docker build -t tantona/sqs-operator .
docker push tantona/sqs-operator:latest

kubectl apply -f ./deploy/crd.yaml
kubectl apply -f ./deploy/operator.yaml

kubectl apply -f ./deploy/cr.yaml
```

## TODO
- CRD validation
