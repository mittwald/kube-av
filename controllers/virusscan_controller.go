package controllers

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	avv1beta1 "github.com/mittwald/kube-av/api/v1beta1"
)

// VirusScanReconciler reconciles a VirusScan object
type VirusScanReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=av.mittwald.de,resources=virusscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=av.mittwald.de,resources=virusscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=av.mittwald.de,resources=virusscans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VirusScanReconciler) Reconcile(ctx context.Context, request ctrl.Request) (res ctrl.Result, err error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Reconciling VirusScan")

	// Fetch the VirusScan instance
	scan := avv1beta1.VirusScan{}

	if err := r.Client.Get(ctx, request.NamespacedName, &scan); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	scanOrig := scan.DeepCopy()
	patch := client.MergeFrom(scanOrig)

	r.observeAndTransition(&scan)

	switch scan.Status.Phase {
	case avv1beta1.VirusScanPhaseUnspecified:
		r.transition(&scan, avv1beta1.VirusScanPhasePending, "")
	case avv1beta1.VirusScanPhasePending:
		res, err = r.reconcilePending(ctx, &scan, reqLogger)
	}

	if !reflect.DeepEqual(&scanOrig.Status, &scan.Status) {
		if err := r.Client.Status().Patch(ctx, &scan, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	return
}

func (r *VirusScanReconciler) observeAndTransition(s *avv1beta1.VirusScan) {
	if s.Status.GetCondition(avv1beta1.VirusScanConditionTypeCompleted) == corev1.ConditionTrue {
		switch s.Status.GetCondition(avv1beta1.VirusScanConditionTypePositive) {
		case corev1.ConditionTrue:
			s.Status.Summary = fmt.Sprintf("Completed (%d infected files)", len(s.Status.ScanResults))
			r.transition(s, avv1beta1.VirusScanPhaseCompletedPositive, "")
		case corev1.ConditionFalse:
			s.Status.Summary = "Completed (no infected files)"
			r.transition(s, avv1beta1.VirusScanPhaseCompletedNegative, "")
		}

		return
	}

	if s.Status.GetCondition(avv1beta1.VirusScanConditionTypeScheduled) == corev1.ConditionTrue {
		s.Status.Summary = "Running"
		r.transition(s, avv1beta1.VirusScanPhaseRunning, "")
		return
	}

	s.Status.Summary = "Pending"
	r.transition(s, avv1beta1.VirusScanPhasePending, "")
}

func (r *VirusScanReconciler) transition(s *avv1beta1.VirusScan, phase avv1beta1.VirusScanPhase, msg string) {
	if s.Status.Phase == phase {
		return
	}

	if msg == "" {
		msg = fmt.Sprintf("transitioning to %s", phase)
	} else {
		msg = fmt.Sprintf("transitioning to %s: %s", phase, msg)
	}

	evtType := corev1.EventTypeNormal

	if phase == avv1beta1.VirusScanPhaseFailed {
		evtType = corev1.EventTypeWarning
	}

	r.Recorder.Eventf(s, evtType, string(phase), msg)
	s.Status.Phase = phase
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirusScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&avv1beta1.VirusScan{}).
		Complete(r)
}
