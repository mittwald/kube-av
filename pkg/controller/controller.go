package controller

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, record.EventRecorder) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, r record.EventRecorder) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, r); err != nil {
			return err
		}
	}
	return nil
}
