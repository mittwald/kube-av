package controllers

import (
	"context"
	"sort"

	"github.com/go-logr/logr"
	avv1beta1 "github.com/mittwald/kube-av/apis/av/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ScheduledVirusScanReconciler) runGarbageCollection(ctx context.Context, svs *avv1beta1.ScheduledVirusScan, l logr.Logger) error {
	vs := avv1beta1.VirusScanList{}

	if err := r.Client.List(ctx, &vs, client.InNamespace(svs.Namespace), client.MatchingLabels{labelScheduledBy: svs.Name}); err != nil {
		return err
	}

	limit := 5
	if l := svs.Spec.HistorySize; l != nil {
		limit = *l
	}

	deletionCandidates := make(scanList, 0, len(vs.Items))
	for i := range vs.Items {
		if vs.Items[i].Status.GetCondition(avv1beta1.VirusScanConditionTypeCompleted) == corev1.ConditionTrue {
			deletionCandidates = append(deletionCandidates, vs.Items[i])
		}
	}

	l.Info("determined candidates for garbage collection", "candidate.count", len(deletionCandidates))

	if len(deletionCandidates) <= limit {
		return nil
	}

	sort.Sort(sort.Reverse(deletionCandidates))

	toDelete := deletionCandidates[limit:]
	for i := range toDelete {
		l.Info("deleting VirusScan", "virusscan.name", toDelete[i].Name)
		if err := r.Client.Delete(ctx, &toDelete[i]); err != nil {
			return err
		}
	}

	return nil
}
