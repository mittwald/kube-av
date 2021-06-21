package controllers

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	avv1beta1 "github.com/mittwald/kube-av/apis/v1beta1"
)

type CronEntry struct {
	entryID  cron.EntryID
	schedule string
}

const labelScheduledBy = "kubeav.mittwald.de/scheduled-by"

// ScheduledVirusScanReconciler reconciles a ScheduledVirusScan object
type ScheduledVirusScanReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	Cron        *cron.Cron
	CronEntries map[string]CronEntry
}

//+kubebuilder:rbac:groups=av.mittwald.de,resources=scheduledvirusscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=av.mittwald.de,resources=scheduledvirusscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=av.mittwald.de,resources=scheduledvirusscans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ScheduledVirusScanReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Reconciling ScheduledVirusScan")

	svs := avv1beta1.ScheduledVirusScan{}

	if err := r.Client.Get(ctx, request.NamespacedName, &svs); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if err := r.runGarbageCollection(ctx, &svs, reqLogger); err != nil {
		return reconcile.Result{}, err
	}

	entry, ok := r.CronEntries[request.NamespacedName.String()]
	if ok && entry.schedule == svs.Spec.Schedule {
		return reconcile.Result{}, nil
	}

	r.Cron.Remove(entry.entryID)

	newEntryID, err := r.Cron.AddFunc(svs.Spec.Schedule, func() {
		svs := avv1beta1.ScheduledVirusScan{}
		ctx := context.Background()

		if err := r.Client.Get(ctx, request.NamespacedName, &svs); err != nil {
			reqLogger.Error(err, "error while re-loading ScheduledVirusScan")
			return
		}

		patch := client.MergeFrom(svs.DeepCopy())

		labels := svs.Spec.Template.Labels
		if labels == nil {
			labels = make(map[string]string)
		}

		labels[labelScheduledBy] = svs.Name

		vs := avv1beta1.VirusScan{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", svs.Name),
				Namespace:    svs.Namespace,
				Labels:       labels,
				Annotations:  svs.Spec.Template.Annotations,
			},
			Spec: svs.Spec.Template.Spec,
		}

		if err := controllerutil.SetControllerReference(&svs, &vs, r.Scheme); err != nil {
			reqLogger.Error(err, "error while setting OwnerReference on VirusScan")
			return
		}

		if err := r.Client.Create(ctx, &vs); err != nil {
			reqLogger.Error(err, "error while creating VirusScan")
			return
		}

		ref, err := reference.GetReference(r.Scheme, &vs)
		if err != nil {
			reqLogger.Error(err, "error while building object reference")
			return
		}

		now := metav1.Now()

		svs.Status.LastScheduledScan = ref
		svs.Status.LastScheduledTime = &now

		if err := r.Client.Status().Patch(ctx, &svs, patch); err != nil {
			reqLogger.Error(err, "error while PATCH'ing ScheduledVirusScan")
			return
		}
	})

	if err != nil {
		return reconcile.Result{}, err
	}

	r.CronEntries[request.NamespacedName.String()] = CronEntry{entryID: newEntryID, schedule: svs.Spec.Schedule}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduledVirusScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&avv1beta1.ScheduledVirusScan{}).
		Complete(r)
}
