package controllers

import (
	"context"

	avv1beta1 "github.com/mittwald/kube-av/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const defaultServiceAccountName = "kubeav"
const defaultClusterRoleName = "kubeav:agent"

func (r *VirusScanReconciler) assertServiceAccount(ctx context.Context, s *avv1beta1.VirusScan) (*corev1.ServiceAccount, error) {
	name := s.Spec.ServiceAccountName
	if name == "" {
		name = defaultServiceAccountName
	}

	serviceAccount := corev1.ServiceAccount{}
	serviceAccountName := types.NamespacedName{Name: name, Namespace: s.Namespace}

	if err := r.Client.Get(ctx, serviceAccountName, &serviceAccount); err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceAccount.Name = serviceAccountName.Name
	serviceAccount.Namespace = serviceAccountName.Namespace

	if err := r.upsert(ctx, &serviceAccount); err != nil {
		return nil, err
	}

	return &serviceAccount, nil
}

func (r *VirusScanReconciler) assertRoleBinding(ctx context.Context, a *corev1.ServiceAccount) (*rbacv1.RoleBinding, error) {
	roleBinding := rbacv1.RoleBinding{}
	roleBindingName := types.NamespacedName{Name: a.Name, Namespace: a.Namespace}

	if err := r.Client.Get(ctx, roleBindingName, &roleBinding); err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	roleBinding.Name = roleBindingName.Name
	roleBinding.Namespace = roleBindingName.Namespace
	roleBinding.Subjects = []rbacv1.Subject{
		{
			Namespace: a.Namespace,
			Name:      a.Name,
			Kind:      "ServiceAccount",
		},
	}
	roleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "ClusterRole",
		Name:     defaultClusterRoleName,
	}

	if err := r.upsert(ctx, &roleBinding); err != nil {
		return nil, err
	}

	return &roleBinding, nil
}
