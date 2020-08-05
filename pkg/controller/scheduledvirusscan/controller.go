package scheduledvirusscan

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"

	avv1beta1 "github.com/mittwald/kube-av/pkg/apis/av/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_scheduledvirusscan")

// Add creates a new ScheduledVirusScan Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, rec record.EventRecorder, cr *cron.Cron) error {
	return add(mgr, newReconciler(mgr, rec, cr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, rec record.EventRecorder, cr *cron.Cron) reconcile.Reconciler {
	return &ReconcileScheduledVirusScan{
		mgr.GetClient(),
		mgr.GetScheme(),
		rec,
		cr,
		make(map[string]cron.EntryID),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("scheduledvirusscan-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ScheduledVirusScan
	err = c.Watch(&source.Kind{Type: &avv1beta1.ScheduledVirusScan{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner ScheduledVirusScan
	err = c.Watch(&source.Kind{Type: &avv1beta1.VirusScan{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &avv1beta1.ScheduledVirusScan{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileScheduledVirusScan implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileScheduledVirusScan{}

// ReconcileScheduledVirusScan reconciles a ScheduledVirusScan object
type ReconcileScheduledVirusScan struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	recorder    record.EventRecorder
	cron        *cron.Cron
	cronEntries map[string]cron.EntryID
}

// Reconcile reads that state of the cluster for a ScheduledVirusScan object and makes changes based on the state read
// and what is in the ScheduledVirusScan.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileScheduledVirusScan) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ScheduledVirusScan")

	ctx := context.Background()
	svs := avv1beta1.ScheduledVirusScan{}

	if err := r.client.Get(ctx, request.NamespacedName, &svs); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	entryID, ok := r.cronEntries[request.NamespacedName.String()]
	if ok {
		r.cron.Remove(entryID)
	}

	newEntryID, err := r.cron.AddFunc(svs.Spec.Schedule, func() {
		svs := avv1beta1.ScheduledVirusScan{}
		ctx := context.Background()

		if err := r.client.Get(ctx, request.NamespacedName, &svs); err != nil {
			reqLogger.Error(err, "error while re-loading ScheduledVirusScan")
			return
		}

		patch := client.MergeFrom(svs.DeepCopy())

		vs := avv1beta1.VirusScan{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", svs.Name),
				Namespace:    svs.Namespace,
				Labels:       svs.Spec.Template.Labels,
				Annotations:  svs.Spec.Template.Annotations,
			},
			Spec: svs.Spec.Template.Spec,
		}

		if err := r.client.Create(ctx, &vs); err != nil {
			reqLogger.Error(err, "error while creating VirusScan")
			return
		}

		ref, err := reference.GetReference(r.scheme, &vs)
		if err != nil {
			reqLogger.Error(err, "error while building object reference")
			return
		}

		now := metav1.Now()

		svs.Status.LastScheduledScan = ref
		svs.Status.LastScheduledTime = &now

		if err := r.client.Status().Patch(ctx, &svs, patch); err != nil {
			reqLogger.Error(err, "error while PATCH'ing ScheduledVirusScan")
			return
		}
	})

	if err != nil {
		return reconcile.Result{}, err
	}

	r.cronEntries[request.NamespacedName.String()] = newEntryID

	return reconcile.Result{}, nil
}
