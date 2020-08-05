package engine

import (
	avv1beta1 "github.com/mittwald/kube-av/pkg/apis/av/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type clamAVEngine struct{}

func (c *clamAVEngine) Name() string {
	return string(avv1beta1.ScanEngineClamAV)
}

func (c *clamAVEngine) AdviseJob(job *batchv1.Job) error {
	container := &job.Spec.Template.Spec.Containers[0]

	container.Image = "quay.io/mittwald/kubeav-agent-clamav:v1"
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "clamdb",
		MountPath: "/var/lib/clamav",
		ReadOnly:  true,
	})

	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: "clamdb",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/lib/clamav",
				},
			},
		},
	)

	return nil
}
