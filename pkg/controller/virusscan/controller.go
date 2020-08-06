package virusscan

import (
	"context"
	"fmt"
	"reflect"

	avv1beta1 "github.com/mittwald/kube-av/pkg/apis/av/v1beta1"
	"github.com/robfig/cron/v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_virusscan")

// Add creates a new VirusScan Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, rec record.EventRecorder, _ *cron.Cron) error {
	return add(mgr, newReconciler(mgr, rec))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, rec record.EventRecorder) reconcile.Reconciler {
	return &ReconcileVirusScan{mgr.GetClient(), mgr.GetScheme(), rec}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("virusscan-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VirusScan
	err = c.Watch(&source.Kind{Type: &avv1beta1.VirusScan{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner VirusScan
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &avv1beta1.VirusScan{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVirusScan implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVirusScan{}

// ReconcileVirusScan reconciles a VirusScan object
type ReconcileVirusScan struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a VirusScan object and makes changes based on the state read
// and what is in the VirusScan.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVirusScan) Reconcile(request reconcile.Request) (res reconcile.Result, err error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VirusScan")

	ctx := context.Background()

	// Fetch the VirusScan instance
	scan := avv1beta1.VirusScan{}

	if err := r.client.Get(ctx, request.NamespacedName, &scan); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
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
		if err := r.client.Status().Patch(ctx, &scan, patch); err != nil {
			return reconcile.Result{}, err
		}
	}

	return
}

func (r *ReconcileVirusScan) observeAndTransition(s *avv1beta1.VirusScan) {
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

func (r *ReconcileVirusScan) transition(s *avv1beta1.VirusScan, phase avv1beta1.VirusScanPhase, msg string) {
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

	r.recorder.Eventf(s, evtType, string(phase), msg)

	s.Status.Phase = phase
}
