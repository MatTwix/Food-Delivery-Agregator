package scheduler

import "github.com/robfig/cron/v3"

type job interface {
	Run()
	Spec() string
}

func RegisterJobs(c *cron.Cron, jobs ...job) {
	for _, job := range jobs {
		c.AddJob(job.Spec(), job)
	}
}
