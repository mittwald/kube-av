package virusscan

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *ReconcileVirusScan) upsert(ctx context.Context, obj runtime.Object) error {
	err := r.client.Update(ctx, obj)
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		return r.client.Create(ctx, obj)
	}

	return err
}
