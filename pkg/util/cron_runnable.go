package util

import (
	"context"

	"github.com/robfig/cron/v3"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ manager.Runnable = &CronRunnable{}

type CronRunnable struct {
	cron *cron.Cron
}

func NewCronRunnable(c *cron.Cron) *CronRunnable {
	return &CronRunnable{c}
}

func (c *CronRunnable) Start(ctx context.Context) error {
	c.cron.Start()

	<-ctx.Done()
	stopCtx := c.cron.Stop()

	<-stopCtx.Done()
	return nil
}

func (c *CronRunnable) NeedLeaderElection() bool {
	return true
}
