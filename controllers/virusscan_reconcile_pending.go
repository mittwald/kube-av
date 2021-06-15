package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	avv1beta1 "github.com/mittwald/kube-av/api/v1beta1"
	"github.com/mittwald/kube-av/pkg/engine"
	"github.com/mittwald/kube-av/pkg/labels"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *VirusScanReconciler) reconcilePending(ctx context.Context, s *avv1beta1.VirusScan, l logr.Logger) (reconcile.Result, error) {
	eng, err := engine.ByName(s.Spec.Engine)
	if err != nil {
		return reconcile.Result{}, err
	}

	sa, err := r.assertServiceAccount(ctx, s)
	if err != nil {
		return reconcile.Result{}, err
	}

	if _, err = r.assertRoleBinding(ctx, sa); err != nil {
		return reconcile.Result{}, err
	}

	job := batchv1.Job{}
	jobName := types.NamespacedName{Name: s.Name, Namespace: s.Namespace}

	if err := r.Client.Get(ctx, jobName, &job); err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	container := corev1.Container{
		Name: "av",
		Args: []string{
			"--engine", eng.Name(),
			"--scan-ref", fmt.Sprintf("%s/%s", s.Namespace, s.Name),
		},
		VolumeMounts: []corev1.VolumeMount{},
	}

	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}

	job.Name = s.Name
	job.Namespace = s.Namespace
	job.Labels[labels.KubernetesManagedBy] = "kubeav"

	job.Spec.Template.Spec.Containers = []corev1.Container{container}
	job.Spec.Template.Spec.ServiceAccountName = sa.Name
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	job.Spec.Template.Spec.Volumes = []corev1.Volume{}

	for i := range s.Spec.Targets {
		volumeName := fmt.Sprintf("scan-target-%d", i)
		relativeMountPath := s.Spec.Targets[i].MountPath

		if relativeMountPath == "" {
			relativeMountPath = volumeName
		}

		absoluteMountPath := fmt.Sprintf("/scan/%s", relativeMountPath)

		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
			Name:         volumeName,
			VolumeSource: s.Spec.Targets[i].Volume,
		})

		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			SubPath:   s.Spec.Targets[i].SubPath,
			MountPath: absoluteMountPath,
		})

		job.Spec.Template.Spec.Containers[0].Args = append(
			job.Spec.Template.Spec.Containers[0].Args,
			"--scan-dir",
			absoluteMountPath,
		)
	}

	if err := eng.AdviseJob(&job); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(s, &job, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	l.Info("upserting batch job")
	if err := r.upsert(ctx, &job); err != nil {
		return reconcile.Result{}, err
	}

	l.Info("PATCHing Condition into VirusScan")
	if err := r.Client.Status().Patch(ctx, s, &PatchVirusScanCondition{
		avv1beta1.VirusScanCondition{
			Type:   avv1beta1.VirusScanConditionTypeScheduled,
			Status: corev1.ConditionTrue,
			Reason: "Scheduled",
		},
	}); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
