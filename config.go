package main

import (
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Job      JobConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Path string
}

type JobConfig struct {
	Name        string
	Description string
	Group       []JobGroupConfig
}

type JobGroupConfig struct {
	Name        string
	Description string
	Task        []JobGroupTaskConfig
}

type JobGroupTaskConfig struct {
	Name        string
	Description string
	Timing      string
	Command     string
}

func OverlapCheck(namelist []string) {
	m := make(map[string]struct{})
	for _, ele := range namelist {
		m[ele] = struct{}{}
	}

	uniq := []string{}
	for i := range m {
		uniq = append(uniq, i)
	}

	if len(namelist) != len(uniq) {
		msg := "name can not use overlap"
		Logger("fatal", "baku.config", msg)
	}
}

func SpaceCheck(name string) {
	check := strings.Contains(name, " ")
	if check == true {
		msg := "name can not use spaces"
		Logger("fatal", "baku.config", msg)
	}
}

func UpperCheck(name string) {
	for _, oneletter := range name {
		check := unicode.IsLower(oneletter)
		if check == false {
			msg := "name can not use upper"
			Logger("fatal", "baku.config", msg)
		}
	}
}

func LoadConfig(configFile string) Config {
	var config Config
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		Logger("fatal", "baku.config", err.Error())
	}

	msg := "load "
	msg += configFile
	Logger("info", "baku.config", msg)
	
	jobGroup := []string{}
	groupTask := []string{}
	for _, group := range config.Job.Group {
		SpaceCheck(group.Name)
		UpperCheck(group.Name)
		jobGroup = append(jobGroup, group.Name)
		for _, task := range group.Task {
			tagName := strings.Join([]string{group.Name, task.Name}, ".")
			SpaceCheck(task.Name)
			UpperCheck(task.Name)
			groupTask = append(groupTask, tagName)
		}
	}

	OverlapCheck(jobGroup)
	OverlapCheck(groupTask)

	return config
}
