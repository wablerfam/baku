package main

import (
	"os/exec"
	"time"

	"github.com/mattn/go-shellwords"
)

type Process struct {
	TagName string
	JobGroupTaskConfig
	Database
}

func (prc Process) ScheduledExec() {
	preExec := prc.PreExec()
	if preExec == "double" {
		msg := "task skip for double startup "
		Logger("warn", prc.TagName, msg)
		return

	} else if preExec == "abend" {
		msg := "task skip for abend state "
		Logger("warn", prc.TagName, msg)
		return

	} else if preExec == "abort" {
		msg := "task skip for abort state "
		Logger("warn", prc.TagName, msg)
		return
	}

	prc.Exec()
}

func (prc Process) PreExec() string {
	preCheck := prc.Database.CheckStatus(prc.TagName)
	if preCheck.Status == "running" {
		return "double"
	} else if preCheck.Status == "abended" {
		return "abend"
	} else if preCheck.Status == "aborted(running)" {
		return "abort"
	}

	return ""
}

func (prc Process) Exec() {
	var (
		status    string
		execTime  float64
		post      *Post
		execStart time.Time
		execEnd   time.Time
	)

	shellwords, _ := shellwords.Parse(prc.JobGroupTaskConfig.Command)

	var cmd *exec.Cmd
	switch len(shellwords) {
	case 1:
		cmd = exec.Command(shellwords[0])
	default:
		cmd = exec.Command(shellwords[0], shellwords[1:]...)
	}

	execStart = time.Now()

	cmd.Start()
	msg := "task start"
	Logger("info", prc.TagName, msg)

	status = "running"
	post = &Post{
		TagName:       prc.TagName,
		Status:        status,
		ExecTime:      execTime,
		ExecStartTime: execStart,
		ExecEndTime:   execEnd,
		ExecCommand:   cmd,
	}

	prc.Database.ChangeStatus(prc.TagName, post)

	err := cmd.Wait()
	if err != nil {
		msg := "task abend"
		Logger("error", prc.TagName, msg)
		status = "abended"
	} else {
		msg := "task end"
		Logger("info", prc.TagName, msg)
		status = "succeeded"
	}

	execEnd = time.Now()

	post = &Post{
		TagName:       prc.TagName,
		Status:        status,
		ExecTime:      (execEnd.Sub(execStart)).Seconds(),
		ExecStartTime: execStart,
		ExecEndTime:   execEnd,
		ExecCommand:   cmd,
	}

	prc.Database.ChangeStatus(prc.TagName, post)
}
