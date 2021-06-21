package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScanEngine string

const (
	ScanEngineClamAV ScanEngine = "ClamAV"
)

type VirusScanTargetSpec struct {
	Volume corev1.VolumeSource `json:"volume"`

	// +optional
	SubPath string `json:"subPath,omitempty"`

	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// ExcludeFiles is a list of regular expressions that is used for excluding
	// files from scanning.
	// +optional
	ExcludeFiles []string `json:"excludeFiles,omitempty"`

	// ExcludeDirectories is a list of regular expressions that is used for
	// excluding directories from scanning.
	// +optional
	ExcludeDirectories []string `json:"excludeDirectories,omitempty"`
}

// VirusScanSpec defines the desired state of VirusScan
type VirusScanSpec struct {
	// Engine describes which AV engine should be used for the virus scan.
	// May be left empty, in which case the default engine (probably ClamAV)
	// will be used.
	// +optional
	Engine ScanEngine `json:"engine,omitempty"`

	// Resources define the resource requirements that should be allocated for
	// the scanner Pod. May be left empty, in which case KubeAV will use some
	// default resource requirements.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Targets defines which directories to scan.
	Targets []VirusScanTargetSpec `json:"targets"`

	// ServiceAccountName specified the name of the service account that should
	// be used for the actual scanning containers. If left empty, this will
	// fall back to a default value.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

type VirusScanPhase string

const (
	VirusScanPhaseUnspecified       VirusScanPhase = ""
	VirusScanPhasePending           VirusScanPhase = "Pending"
	VirusScanPhaseRunning           VirusScanPhase = "Running"
	VirusScanPhaseCompletedPositive VirusScanPhase = "CompletedPositive"
	VirusScanPhaseCompletedNegative VirusScanPhase = "CompletedNegative"
	VirusScanPhaseFailed            VirusScanPhase = "Failed"
)

type VirusScanResultItem struct {
	FilePath          string `json:"filePath"`
	MatchingSignature string `json:"matchingSignature"`
}

// VirusScanStatus defines the observed state of VirusScan
type VirusScanStatus struct {
	// +optional
	Conditions map[VirusScanConditionType]VirusScanCondition `json:"conditions,omitempty"`

	// +optional
	Phase VirusScanPhase `json:"phase,omitempty"`

	// Summary contains a human-readable summary of the current scan status
	// +optional
	Summary string `json:"summary,omitempty"`

	// Job is a reference to the respective batchv1.Job object as soon as the
	// VirusScan has been scheduled.
	// +optional
	Job *corev1.ObjectReference `json:"job,omitempty"`

	// ScanResults contains the results of the virus scan.
	// +optional
	ScanResults []VirusScanResultItem `json:"scanResults,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=virusscans,scope=Namespaced,shortName=vs
//+kubebuilder:printcolumn:name="Summary",type="string",JSONPath=".status.summary",description="A summary on the current scan phase"
//+kubebuilder:printcolumn:name="Scheduled",type="date",JSONPath=".status.conditions.Scheduled.lastTransitionTime",description="Tells when this scan was scheduled"
//+kubebuilder:printcolumn:name="Completed",type="date",JSONPath=".status.conditions.Completed.lastTransitionTime",description="Tells whether this scan has completed"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// VirusScan is the Schema for the virusscans API
type VirusScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirusScanSpec   `json:"spec,omitempty"`
	Status VirusScanStatus `json:"status,omitempty"`
}

type VirusScanTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirusScanSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VirusScanList contains a list of VirusScan
type VirusScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirusScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirusScan{}, &VirusScanList{})
}
