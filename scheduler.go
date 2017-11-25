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

func (sch Scheduler) Run(groups []JobGroupConfig, database Database) {
	Logger("info", "baku.scheduler", "baku schduler up")

	for _, group := range groups {
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")

			process := Process{tagName, task, database}

			sch.Timer.AddFunc(task.Timing, func() { process.ScheduledExec() })
		}
	}

	sch.Timer.Start()
}
