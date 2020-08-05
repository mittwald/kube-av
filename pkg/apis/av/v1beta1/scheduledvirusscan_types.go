package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScheduledVirusScanSpec defines the desired state of ScheduledVirusScan
type ScheduledVirusScanSpec struct {
	Schedule string            `json:"schedule"`
	Template VirusScanTemplate `json:"template"`
}

// ScheduledVirusScanStatus defines the observed state of ScheduledVirusScan
type ScheduledVirusScanStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledVirusScan is the Schema for the scheduledvirusscans API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scheduledvirusscans,scope=Namespaced
type ScheduledVirusScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledVirusScanSpec   `json:"spec,omitempty"`
	Status ScheduledVirusScanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledVirusScanList contains a list of ScheduledVirusScan
type ScheduledVirusScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledVirusScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledVirusScan{}, &ScheduledVirusScanList{})
}
