package main

import (
	"strings"

	"github.com/robfig/cron"
)

type Scheduler struct {
	Timer *cron.Cron
}

func InitScheduler() *Scheduler {
	s := new(Scheduler)
	s.Timer = cron.New()
	return s
}

func (sch Scheduler) Run(job JobConfig, database Database) {
	msg := "set "
	msg += job.Name
	Logger("info", "baku.scheduler", msg)

	for _, group := range job.Group {
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")

			process := Process{tagName, task, database}

			sch.Timer.AddFunc(task.Timing, func() { process.ScheduledExec() })
		}
	}

	sch.Timer.Start()
}
