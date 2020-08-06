package util

import "github.com/robfig/cron/v3"

type CronRunnable struct {
	cron *cron.Cron
}

func NewCronRunnable(c *cron.Cron) *CronRunnable {
	return &CronRunnable{c}
}

func (c *CronRunnable) Start(i <-chan struct{}) error {
	c.cron.Start()

	<-i
	stopCtx := c.cron.Stop()

	<-stopCtx.Done()
	return nil
}

func (c *CronRunnable) NeedLeaderElection() bool {
	return true
}
