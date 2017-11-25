package main

import (
	"encoding/json"
	_ "log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	JobConfig
	Database
}

type JsonStatus struct {
	JsonJob `json:"job"`
}

type JsonJob struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	JsonJobGroups `json:"groups"`
}

type JsonJobGroup struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	JsonJobGroupTasks `json:"tasks"`
}

type JsonJobGroups []JsonJobGroup

type JsonJobGroupTask struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	ExecTime      float64 `json:"exec_time"`
	ExecStartTime string  `json:"last_start_time"`
	ExecEndTime   string  `json:"last_end_time"`
}

type JsonJobGroupTasks []JsonJobGroupTask

type JsonActions struct {
	JsonActionsMesseages `json:"actions"`
}

type JsonActionsMesseage struct {
	Action      string `json:"action"`
	Url         string `json:"url"`
	Parameter   string `json:"parameter"`
	Description string `json:"description"`
}

type JsonActionsMesseages []JsonActionsMesseage

type JsonActionsResponse struct {
	Action    string `json:"action"`
	GroupName string `json:"groupname"`
	TaskName  string `json:"taskname"`
	Result    string `json:"result"`
	Message   string `json:"message"`
}

func (hdl Handler) Use(router *mux.Router) {
	router.HandleFunc("/", hdl.DefaultHandler)
	router.HandleFunc("/ping", hdl.PingHandler).Methods("GET")
	router.HandleFunc("/api/status", hdl.StatusHandler).Methods("GET")
	router.HandleFunc("/api/actions", hdl.ActionsHandler).Methods("GET")
	router.HandleFunc("/api/actions/exec", hdl.ActionsExecHandler).Methods("POST")
	router.HandleFunc("/api/actions/kill", hdl.ActionsKillHandler).Methods("POST")
	router.HandleFunc("/api/actions/refresh", hdl.ActionsRefreshHandler).Methods("POST")
}

func (hdl Handler) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("my name is baku"))
}

func (hdl Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (hdl Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	var jsonGroups JsonJobGroups
	for _, group := range hdl.JobConfig.Group {
		var jsonTasks JsonJobGroupTasks
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")
			stat := hdl.Database.CheckStatus(tagName)

			execTime := stat.ExecTime
			if stat.Status == "running" {
				execTime = (time.Now().Sub(stat.ExecStartTime)).Seconds()
			}

			startTime := stat.ExecStartTime.String()
			if stat.ExecStartTime.IsZero() == true {
				startTime = ""
			}

			endTime := stat.ExecEndTime.String()
			if stat.ExecEndTime.IsZero() == true {
				endTime = ""
			}

			setTask := JsonJobGroupTask{task.Name, task.Description, stat.Status, execTime, startTime, endTime}
			jsonTasks = append(jsonTasks, setTask)
		}
		setGroup := JsonJobGroup{group.Name, group.Description, jsonTasks}
		jsonGroups = append(jsonGroups, setGroup)
	}

	jsonJob := JsonJob{hdl.JobConfig.Name, hdl.JobConfig.Description, jsonGroups}
	jsonStatus := JsonStatus{jsonJob}

	json.NewEncoder(w).Encode(jsonStatus)
}

func (hdl Handler) ActionsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		action      string
		url         string
		parameter   string
		description string

		jsonMesseage  JsonActionsMesseage
		jsonMesseages JsonActionsMesseages
	)

	action = "exec"
	url = "api/actions/exec"
	parameter = "?job=[job_name]&group=[group_name]&task=[task_name]"
	description = "execute task immediately"

	jsonMesseage = JsonActionsMesseage{action, url, parameter, description}

	jsonMesseages = append(jsonMesseages, jsonMesseage)

	action = "kill"
	url = "api/actions/kill"
	parameter = "?job=[job_name]&group=[group_name]&task=[task_name]"
	description = "kill task immediately"

	jsonMesseage = JsonActionsMesseage{action, url, parameter, description}

	jsonMesseages = append(jsonMesseages, jsonMesseage)

	action = "refresh"
	url = "api/actions/refresh"
	parameter = "?job=[job_name]&group=[group_name]&task=[task_name]"
	description = "abended or killed or aborted schedule refreshes the task of unexecuted status"

	jsonMesseage = JsonActionsMesseage{action, url, parameter, description}

	jsonMesseages = append(jsonMesseages, jsonMesseage)

	jsonActions := JsonActions{jsonMesseages}

	json.NewEncoder(w).Encode(jsonActions)
}

func (hdl Handler) PreActionsExec(query map[string][]string, actionType string) JsonActionsResponse {
	var (
		action    string = actionType
		groupName string
		taskName  string
		result    string
		message   string

		res JsonActionsResponse
	)

	if len(query) == 0 {
		result = "failed"
		message = "could not execute because parameter does not exist"

		res = JsonActionsResponse{Action: action, Result: result, Message: message}

		return res
	}

	if _, ok := query["group"]; ok {
	} else {
		result = "failed"
		message = "need to specify the group parameter"

		res = JsonActionsResponse{Action: action, Result: result, Message: message}

		return res
	}

	if _, ok := query["task"]; ok {
	} else {
		result = "failed"
		message = "need to specify the task parameter"

		res = JsonActionsResponse{Action: action, Result: result, Message: message}

		return res
	}

	tag := strings.Join([]string{query["group"][0], query["task"][0]}, ".")

	preCheck := hdl.Database.CheckStatus(tag)
	if preCheck.TagName == "" {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "incorrect group or task"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		return res
	}

	return res
}

func (hdl Handler) ActionsExecHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var (
		action    string = "exec"
		groupName string
		taskName  string
		result    string
		message   string

		res JsonActionsResponse
	)

	preCheck := hdl.PreActionsExec(query, action)
	if preCheck.Result == "failed" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(preCheck)
		return
	}

	tag := strings.Join([]string{query["group"][0], query["task"][0]}, ".")

	var taskConfig JobGroupTaskConfig

	for _, group := range hdl.JobConfig.Group {
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")
			if tagName == tag {
				taskConfig = task
			}
		}
	}

	process := Process{TagName: tag, JobGroupTaskConfig: taskConfig, Database: hdl.Database}

	preExec := process.PreExec()
	if preExec == "double" {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "double check failed"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(res)
		return

	} else if preExec == "abend" {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "abend check failed"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(res)
		return

	} else if preExec == "abort" {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "abort check failed"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(res)
		return
	}

	go func() {
		process.Exec()
	}()

	groupName = query["group"][0]
	taskName = query["task"][0]
	result = "success"

	res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(res)
}

func (hdl Handler) ActionsKillHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var (
		action    string = "kill"
		groupName string
		taskName  string
		result    string
		message   string

		res JsonActionsResponse
	)

	preCheck := hdl.PreActionsExec(query, action)
	if preCheck.Result == "failed" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(preCheck)
		return
	}

	tag := strings.Join([]string{query["group"][0], query["task"][0]}, ".")

	runningCheck := hdl.Database.CheckStatus(tag)
	if runningCheck.Status != "running" {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "kill process is not runnnig state"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(res)
		return
	}

	runCmd := hdl.Database.CheckStatus(tag).ExecCommand
	runCmd.Process.Kill()

	groupName = query["group"][0]
	taskName = query["task"][0]
	result = "success"

	res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(res)
}

func (hdl Handler) ActionsRefreshHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var (
		action    string = "refresh"
		groupName string
		taskName  string
		result    string
		message   string

		res JsonActionsResponse

		post *Post
	)

	preCheck := hdl.PreActionsExec(query, action)
	if preCheck.Result == "failed" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(preCheck)
		return
	}

	tag := strings.Join([]string{query["group"][0], query["task"][0]}, ".")

	refreshCheck := hdl.Database.CheckStatus(tag)
	if (refreshCheck.Status != "abended") && (refreshCheck.Status != "aborted(running)") {
		groupName = query["group"][0]
		taskName = query["task"][0]
		result = "failed"
		message = "refresh can be used only in the abend and abort state"

		res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(res)
		return
	}

	post = &Post{
		TagName:       refreshCheck.TagName,
		Status:        "refreshed",
		ExecTime:      refreshCheck.ExecTime,
		ExecStartTime: refreshCheck.ExecStartTime,
		ExecEndTime:   refreshCheck.ExecEndTime,
		ExecCommand:   refreshCheck.ExecCommand,
	}

	hdl.Database.ChangeStatus(tag, post)

	groupName = query["group"][0]
	taskName = query["task"][0]
	result = "success"
	message = ""

	res = JsonActionsResponse{Action: action, GroupName: groupName, TaskName: taskName, Result: result, Message: message}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(res)
}
