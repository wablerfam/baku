package main

import (
	"flag"
)

func main() {
	Logger("info", "baku.main", "start")

	var (
		parseFile = flag.String("c", "default", "specify config file")
	)

	flag.Parse()

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
