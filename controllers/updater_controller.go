package controllers

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/mittwald/kube-av/pkg/engine"

	"github.com/go-logr/logr"
	"github.com/mittwald/kube-av/pkg/labels"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ manager.Runnable = &UpdaterController{}

type UpdaterController struct {
	client            client.Client
	logger            logr.Logger
	operatorNamespace string
}

func NewUpdaterController(c client.Client, l logr.Logger, ns string) *UpdaterController {
	return &UpdaterController{
		c,
		l,
		ns,
	}
}

func (u *UpdaterController) Start(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Minute)
	retry := make(chan struct{})

	reconcileOrRetry := func() {
		//goland:noinspection GoNilness
		if res, err := u.reconcile(ctx); res.Requeue || err != nil {
			if err != nil {
				u.logger.Error(err, "error while reconciling updater daemonset")
			}

			after := res.RequeueAfter
			if after == 0 {
				after = 10 * time.Second
			}

			go func() {
				time.Sleep(after)
				retry <- struct{}{}
			}()
		}
	}

	go func() {
		retry <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			reconcileOrRetry()
		case <-retry:
			reconcileOrRetry()
		}
	}
}

func (u *UpdaterController) reconcile(ctx context.Context) (ctrl.Result, error) {
	ds := appsv1.DaemonSet{}
	dsName := types.NamespacedName{Namespace: u.operatorNamespace, Name: "kubeav-updater"}

	cm := corev1.ConfigMap{}
	cmName := types.NamespacedName{Namespace: u.operatorNamespace, Name: dsName.Name}

	u.logger.Info("reconciling updater DaemonSet")

	if err := u.client.Get(ctx, dsName, &ds); err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err := u.client.Get(ctx, cmName, &cm); err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if cm.Labels == nil {
		cm.Labels = make(map[string]string)
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	cm.Namespace = cmName.Namespace
	cm.Name = cmName.Name
	cm.Labels[labels.KubernetesManagedBy] = "kubeav"
	cm.Data["freshclam.conf"] = freshclamTemplate

	if err := u.upsert(ctx, &cm); err != nil {
		return reconcile.Result{}, err
	}

	t := true

	i := corev1.Container{
		Name:    "chown",
		Image:   "busybox",
		Command: []string{"sh", "-c", "chown 100:101 /var/lib/clamav"},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "clamdb",
			MountPath: "/var/lib/clamav",
		}},
		SecurityContext: &corev1.SecurityContext{
			Privileged: &t,
		},
	}

	c := corev1.Container{}
	c.Name = "kubeav-updater"
	c.Image = engine.ClamavUpdaterImage
	c.ImagePullPolicy = corev1.PullIfNotPresent
	c.VolumeMounts = []corev1.VolumeMount{{
		Name:      "clamdb",
		MountPath: "/var/lib/clamav",
	}, {
		Name:      "clamconfig",
		MountPath: "/etc/kubeav",
	}}
	c.Args = []string{"--config-file", "/etc/kubeav/freshclam.conf"}
	c.Resources = corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    engine.ClamavUpdaterCPURequest.Quantity,
			corev1.ResourceMemory: engine.ClamavUpdaterMemoryRequest.Quantity,
		},
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    engine.ClamavUpdaterCPULimit.Quantity,
			corev1.ResourceMemory: engine.ClamavUpdaterMemoryLimit.Quantity,
		},
	}

	ds.Name = dsName.Name
	ds.Namespace = dsName.Namespace
	ds.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			labels.KubernetesAppName:      "kubeav",
			labels.KubernetesAppComponent: "updater",
		},
	}
	ds.Spec.Template.Labels = map[string]string{
		labels.KubernetesAppName:      "kubeav",
		labels.KubernetesAppComponent: "updater",
		labels.KubernetesManagedBy:    "kubeav",
	}
	ds.Spec.Template.Spec.Volumes = []corev1.Volume{{
		Name: "clamdb",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: engine.ClamavLibraryHostPath,
			},
		},
	}, {
		Name: "clamconfig",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.Name,
				},
			},
		},
	}}
	ds.Spec.Template.Spec.InitContainers = []corev1.Container{i}
	ds.Spec.Template.Spec.Containers = []corev1.Container{c}

	return ctrl.Result{}, u.upsert(ctx, &ds)
}

func (u *UpdaterController) NeedLeaderElection() bool {
	return true
}

func (u *UpdaterController) upsert(ctx context.Context, obj client.Object) error {
	err := u.client.Update(ctx, obj)
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		return u.client.Create(ctx, obj)
	}

	return err
}
