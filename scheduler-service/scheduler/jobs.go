package scheduler

import "github.com/robfig/cron/v3"

type job interface {
	Run()
}

func RegisterJobs(c *cron.Cron, jobs ...job) {
	for _, job := range jobs {
		c.AddJob("@every 30s", job)
	}
}
