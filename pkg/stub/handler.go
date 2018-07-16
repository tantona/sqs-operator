package stub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
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
	log.SetLevel(log.DebugLevel)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	h.SQSClient = sqs.New(sess)
	switch o := event.Object.(type) {
	case *v1.SQSQueue:
		sqsqueue := o
		if event.Deleted {
			return h.deleteSQSQueue(sqsqueue)
		}

		if !h.queueExists(sqsqueue) {
			if err := h.createSQSQueue(sqsqueue); err != nil {
				return err
			}
			return nil
		}

		hasChanged, err := h.queueHasChanged(sqsqueue)
		if err != nil {
			return err
		}

		if hasChanged {
			if err := h.updateSQSQueue(sqsqueue); err != nil {
				return err
			}
		}

		if err := sdk.Update(sqsqueue); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) queueExists(cr *v1.SQSQueue) bool {
	_, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist {
			return false
		}
	}

	return true
}

func (h *Handler) deleteSQSQueue(cr *v1.SQSQueue) error {
	getQueueURLResponse, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})

	if err != nil {
		return err
	}

	_, err = h.SQSClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: getQueueURLResponse.QueueUrl,
	})

	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) createSQSQueue(cr *v1.SQSQueue) error {
	createQueueResponse, err := h.SQSClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName:  aws.String(cr.Spec.Name),
		Attributes: buildAttributes(cr),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == sqs.ErrCodeQueueDeletedRecently {
				log.Info("queue was deleted recently attempting to create in 60 seconds")
				time.Sleep(60 * time.Second)
				return h.createSQSQueue(cr)
			}
		}
		return err

	}

	log.Infof("sqs queue created: %s", *createQueueResponse.QueueUrl)
	return h.setSQSQueueAnnotations(cr)
}

func (h *Handler) updateSQSQueue(cr *v1.SQSQueue) error {
	r, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})
	if err != nil {
		return err
	}

	if _, err := h.SQSClient.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		QueueUrl:   r.QueueUrl,
		Attributes: buildAttributes(cr),
	}); err != nil {
		return err
	}

	log.Infof("sqs queue updated: %s", *r.QueueUrl)
	return h.setSQSQueueAnnotations(cr)
}

func (h *Handler) getQueueAttributes(cr *v1.SQSQueue) (map[string]string, error) {
	attrs := map[string]string{}
	getQueueURLResponse, err := h.SQSClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(cr.Spec.Name),
	})

	if err != nil {
		return attrs, err
	}

	attrs["QueueUrl"] = *getQueueURLResponse.QueueUrl

	getQueueAttributesResponse, err := h.SQSClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       getQueueURLResponse.QueueUrl,
		AttributeNames: []*string{aws.String("All")},
	})
	if err != nil {
		return attrs, err
	}

	for k, v := range getQueueAttributesResponse.Attributes {
		attrs[k] = *v
	}

	return attrs, nil
}

func (h *Handler) queueHasChanged(cr *v1.SQSQueue) (bool, error) {
	attrs, err := h.getQueueAttributes(cr)
	if err != nil {
		return false, err
	}
	stringAttrs := []string{"VisibilityTimeout", "MaximumMessageSize", "MessageRetentionPeriod", "DelaySeconds", "ReceiveMessageWaitTimeSeconds"}
	for _, key := range stringAttrs {
		if cr.Spec.Attributes[key] != attrs[key] {
			log.Debugf("%s: %s != %s", key, cr.Spec.Attributes[key], attrs[key])
			return true, nil
		}
	}

	equal, err := compareJSON(cr.Spec.Attributes["RedrivePolicy"], attrs["RedrivePolicy"])
	if !equal {
		log.Debugf("RedrivePolicy not equal %s %s", stripWhitespace(cr.Spec.Attributes["RedrivePolicy"]), attrs["RedrivePolicy"])
		return true, nil
	}

	return false, nil
}

func (h *Handler) setSQSQueueAnnotations(cr *v1.SQSQueue) error {
	attributes, err := h.getQueueAttributes(cr)
	if err != nil {
		return err
	}

	for key := range attributes {
		cr.Annotations[fmt.Sprintf("%s/%s", annotationPrefix, key)] = attributes[key]
	}
	return nil
}

func buildAttributes(cr *v1.SQSQueue) map[string]*string {
	attributes := map[string]*string{}
	for k, v := range cr.Spec.Attributes {
		attributes[k] = aws.String(stripWhitespace(v))
	}
	return attributes
}

func compareJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(stripWhitespace(s1)), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %v", err)
	}
	err = json.Unmarshal([]byte(stripWhitespace(s2)), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %v", err)
	}

	return reflect.DeepEqual(o1, o2), nil
}

func stripWhitespace(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "", -1), " ", "", -1)
}
