package main

import (
	"flag"
	"fmt"
)

var version string

func chechVerrsion() {
	fmt.Println(version)	
}

func main() {
	var (
		parseVersion = flag.Bool("v", false, "check version")
		parseFile = flag.String("c", "default", "specify config file")
	)

	flag.Parse()

	if *parseVersion == true {
		chechVerrsion()
		return
	}

	Logger("info", "baku.main", "start")
	
	if *parseFile == "default" {
		Logger("fatal", "baku.main", "not set -c [config_file]")
	}

	conf := LoadConfig(*parseFile)

	database := LoadDatabase(conf.Database, "data")
	database.Setup(conf.Job.Group)

	scheduler := InitScheduler()
	scheduler.Run(conf.Job, database)

	server := InitServer(conf.Server)
	server.Run(conf.Job, database)
}
