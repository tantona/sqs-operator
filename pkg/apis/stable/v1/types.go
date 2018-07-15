package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SQSQueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SQSQueue `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SQSQueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SQSQueueSpec   `json:"spec"`
	Status            SQSQueueStatus `json:"status,omitempty"`
}

type RedrivePolicy struct {
	MaxReceiveCount     string `json:"maxReceiveCount"`
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
}

type SQSQueueSpec struct {
	Name                          string        `json:"name"`
	VisibilityTimeout             string        `json:"visibilityTimeout"`
	MaximumMessageSize            string        `json:"maximumMessageSize"`
	MessageRetentionPeriod        string        `json:"messageRetentionPeriod"`
	DelaySeconds                  string        `json:"delaySeconds"`
	ReceiveMessageWaitTimeSeconds string        `json:"receiveMessageWaitTimeSeconds"`
	RedrivePolicy                 RedrivePolicy `json:"redrivePolicy"`
	FifoQueue                     bool          `json:"fifoQueue"`
	// ContentBasedDeduplication     bool          `json:"contentBasedDeduplication"`
	// KmsMasterKeyID                string        `json:"kmsMasterKeyId"`
	// KmsDataKeyReusePeriodSeconds  string        `json:"kmsDataKeyReusePeriodSeconds"`
}
type SQSQueueStatus struct {
	// Fill me
}
