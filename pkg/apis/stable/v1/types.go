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

type SQSQueueSpec struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}
type SQSQueueStatus struct {
	// Fill me
}
