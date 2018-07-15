# SQS Operator

This operator allows you to manage AWS SQS queues with CustomResourceDefinitions in Kubernetes.


## Getting Started

IAM policy 

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

Create Resources
```
kubectl apply -f ./deploy/crd.yaml
kubectl apply -f ./deploy/operator.yaml

kubectl apply -f ./deploy/queue.yaml
```

## TODO
- CRD validation
