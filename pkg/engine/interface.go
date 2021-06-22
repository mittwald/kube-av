package engine

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"

	avv1beta1 "github.com/mittwald/kube-av/apis/av/v1beta1"
)

type ScanEngine interface {
	Name() string
	AdviseJob(job *batchv1.Job) error

	Execute(context.Context, *avv1beta1.VirusScan, []string) (*ScanReport, error)
}

type ErrUnknownEngine avv1beta1.ScanEngine

func (e ErrUnknownEngine) Error() string {
	return fmt.Sprintf("unknown AV scan engine: %s", string(e))
}

type ScanReportItem struct {
	FilePath         string
	MatchedSignature string
}

type ScanReport struct {
	InfectedFiles []ScanReportItem
}

func ByName(e avv1beta1.ScanEngine) (ScanEngine, error) {
	switch e {
	case avv1beta1.ScanEngineClamAV:
		return &clamAVEngine{}, nil
	default:
		return nil, ErrUnknownEngine(e)
	}
}
