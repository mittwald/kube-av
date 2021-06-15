package controllers

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *VirusScanReconciler) upsert(ctx context.Context, obj client.Object) error {
	err := r.Client.Update(ctx, obj)
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		return r.Client.Create(ctx, obj)
	}

	return err
}
