package controllers

import (
	"encoding/json"
	"fmt"

	avv1beta1 "github.com/mittwald/kube-av/apis/av/v1beta1"
	"github.com/mittwald/kube-av/pkg/engine"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type PatchVirusScanCondition struct {
	Condition avv1beta1.VirusScanCondition
}

func (p *PatchVirusScanCondition) Type() types.PatchType {
	return types.MergePatchType
}

func (p *PatchVirusScanCondition) Data(obj runtime.Object) ([]byte, error) {
	cnd := p.Condition
	cnd.LastTransitionTime = metav1.Now()

	patch := map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": map[avv1beta1.VirusScanConditionType]interface{}{
				cnd.Type: cnd,
			},
		},
	}

	return json.Marshal(patch)
}

type PatchVirusScanResult struct {
	ScanReport *engine.ScanReport
}

func (p *PatchVirusScanResult) Type() types.PatchType {
	return types.MergePatchType
}

func (p *PatchVirusScanResult) Data(obj runtime.Object) ([]byte, error) {
	now := metav1.Now()
	conditions := map[avv1beta1.VirusScanConditionType]avv1beta1.VirusScanCondition{
		avv1beta1.VirusScanConditionTypeCompleted: {
			Type:               avv1beta1.VirusScanConditionTypeCompleted,
			Status:             corev1.ConditionTrue,
			Message:            "scan completed",
			LastTransitionTime: now,
		},
	}

	if len(p.ScanReport.InfectedFiles) > 0 {
		conditions[avv1beta1.VirusScanConditionTypePositive] = avv1beta1.VirusScanCondition{
			Type:               avv1beta1.VirusScanConditionTypePositive,
			Status:             corev1.ConditionTrue,
			Message:            fmt.Sprintf("found %d infected files", len(p.ScanReport.InfectedFiles)),
			LastTransitionTime: now,
		}
	} else {
		conditions[avv1beta1.VirusScanConditionTypePositive] = avv1beta1.VirusScanCondition{
			Type:               avv1beta1.VirusScanConditionTypePositive,
			Status:             corev1.ConditionFalse,
			Message:            "found no infected files",
			LastTransitionTime: now,
		}
	}

	results := make([]avv1beta1.VirusScanResultItem, len(p.ScanReport.InfectedFiles))
	for i := range p.ScanReport.InfectedFiles {
		results[i].FilePath = p.ScanReport.InfectedFiles[i].FilePath
		results[i].MatchingSignature = p.ScanReport.InfectedFiles[i].MatchedSignature
	}

	patch := map[string]interface{}{
		"status": map[string]interface{}{
			"conditions":  conditions,
			"scanResults": results,
		},
	}

	return json.Marshal(patch)
}
