package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/tantona/sqs-operator/pkg/apis/stable/v1"
)

const annotationPrefix = "tantona.k8s.operator.sqs"

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	SQSClient sqsiface.SQSAPI
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	h.SQSClient = sqs.New(sess)
	switch o := event.Object.(type) {
	case *v1.SQSQueue:
		sqsqueue := o
		if event.Deleted {
			return h.deleteSQSQueue(sqsqueue)
		}

		if err := h.createSQSQueue(sqsqueue); err != nil {
			return err
		}

		if err := sdk.Update(sqsqueue); err != nil {
			return err
		}

		return nil
	}
	return nil
}

func (h *Handler) deleteSQSQueue(cr *v1.SQSQueue) error {
	getQueueUrlResponse, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})

	if err != nil {
		return err
	}

	_, err = h.SQSClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: getQueueUrlResponse.QueueUrl,
	})

	if err != nil {
		return err
	}

	return nil
}

func buildAttributes(cr *v1.SQSQueue) map[string]*string {
	attributes := map[string]*string{}
	if cr.Spec.RedrivePolicy.MaxReceiveCount != "" {
		b, _ := json.Marshal(cr.Spec.RedrivePolicy)
		attributes["RedrivePolicy"] = aws.String(string(b))
	}

	if cr.Spec.VisibilityTimeout != "" {
		attributes["VisibilityTimeout"] = aws.String(cr.Spec.VisibilityTimeout)
	}

	if cr.Spec.MaximumMessageSize != "" {
		attributes["MaximumMessageSize"] = aws.String(cr.Spec.MaximumMessageSize)
	}

	if cr.Spec.MessageRetentionPeriod != "" {
		attributes["MessageRetentionPeriod"] = aws.String(cr.Spec.MessageRetentionPeriod)
	}

	if cr.Spec.DelaySeconds != "" {
		attributes["DelaySeconds"] = aws.String(cr.Spec.DelaySeconds)
	}

	if cr.Spec.ReceiveMessageWaitTimeSeconds != "" {
		attributes["ReceiveMessageWaitTimeSeconds"] = aws.String(cr.Spec.ReceiveMessageWaitTimeSeconds)
	}

	if cr.Spec.FifoQueue {
		attributes["FifoQueue"] = aws.String("true")
	}

	return attributes
}

func (h *Handler) createSQSQueue(cr *v1.SQSQueue) error {
	attributes := buildAttributes(cr)

	createQueueResponse, err := h.SQSClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName:  aws.String(cr.Spec.Name),
		Attributes: attributes,
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == sqs.ErrCodeQueueDeletedRecently {
				log.Info("queue was deleted recently attempting to create in 60 seconds")
				time.Sleep(60 * time.Second)
				return h.createSQSQueue(cr)
			}

			if awsErr.Code() == sqs.ErrCodeQueueNameExists {
				log.Info("queue exists updating attributes")
				return h.updateSQSQueueAttributes(cr)
			}
		}
		return err

	}

	log.Info("sqs queue created: %s", cr.Annotations[fmt.Sprintf("%s/QueueArn", annotationPrefix)])
	return h.setSQSQueueAnnotations(createQueueResponse.QueueUrl, cr)
}

func (h *Handler) updateSQSQueueAttributes(cr *v1.SQSQueue) error {
	r, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})
	if err != nil {
		return err
	}

	if _, err := h.SQSClient.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(cr.Annotations[fmt.Sprintf("%s/QueueUrl", annotationPrefix)]),
		Attributes: buildAttributes(cr),
	}); err != nil {
		return err
	}

	log.Info("sqs queue updated: %s", cr.Annotations[fmt.Sprintf("%s/QueueArn", annotationPrefix)])
	return h.setSQSQueueAnnotations(r.QueueUrl, cr)
}

func (h *Handler) setSQSQueueAnnotations(queueUrl *string, cr *v1.SQSQueue) error {
	getQueueAttributesResponse, err := h.SQSClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       queueUrl,
		AttributeNames: []*string{aws.String("QueueArn")},
	})
	if err != nil {
		return err
	}

	cr.Annotations[fmt.Sprintf("%s/QueueUrl", annotationPrefix)] = *queueUrl
	cr.Annotations[fmt.Sprintf("%s/QueueArn", annotationPrefix)] = *getQueueAttributesResponse.Attributes["QueueArn"]
	return nil
}
