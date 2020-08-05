package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScheduledVirusScanSpec defines the desired state of ScheduledVirusScan
type ScheduledVirusScanSpec struct {
	// Schedule is a crontab schedule that describes how often a VirusScan
	// resource should be created.
	Schedule string `json:"schedule"`

	// HistorySize is the amount of VirusScan jobs that should be kept after
	// completion.
	// +optional
	HistorySize *int `json:"historySize,omitempty"`

	// Template is the template for the actual VirusScan object.
	Template VirusScanTemplate `json:"template"`
}

// ScheduledVirusScanStatus defines the observed state of ScheduledVirusScan
type ScheduledVirusScanStatus struct {
	// LastScheduledScan is a reference to the last VirusScan object created
	// from this ScheduledVirusScan instance.
	// +optional
	LastScheduledScan *corev1.ObjectReference `json:"lastScheduledScan,omitempty"`

	// LastScheduledTime is the time at which the last VirusScan was scheduled
	// +optional
	LastScheduledTime *metav1.Time `json:"lastScheduledTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduledVirusScan is the Schema for the scheduledvirusscans API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scheduledvirusscans,scope=Namespaced,shortName=svs
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",description="The schedule after which the virus scan should be scheduled"
// +kubebuilder:printcolumn:name="Last scan",type="string",JSONPath=".status.lastScheduledScan.name",description="The latest scan created from this scheduled scan"
// +kubebuilder:printcolumn:name="Last scheduled",type="date",JSONPath=".status.lastScheduledTime",description="Tells when this scan was last scheduled"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
