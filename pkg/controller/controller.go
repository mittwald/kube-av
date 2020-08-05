package controller

import (
	"github.com/robfig/cron/v3"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, record.EventRecorder, *cron.Cron) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, r record.EventRecorder, c *cron.Cron) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, r, c); err != nil {
			return err
		}
	}
	return nil
}
