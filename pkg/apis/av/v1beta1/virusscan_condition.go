package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirusScanConditionType string

const (
	VirusScanConditionTypeScheduled VirusScanConditionType = "Scheduled"
	VirusScanConditionTypeCompleted VirusScanConditionType = "Completed"
	VirusScanConditionTypePositive  VirusScanConditionType = "Positive"
)

type VirusScanCondition struct {
	Type VirusScanConditionType `json:"type"`

	Status corev1.ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

func (in *VirusScanStatus) GetCondition(conditionType VirusScanConditionType) corev1.ConditionStatus {
	cnd, ok := in.Conditions[conditionType]
	if !ok {
		return corev1.ConditionUnknown
	}

	return cnd.Status
}