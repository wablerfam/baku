package main

import (
	"flag"
)

func main() {
	Logger("info", "baku.main", "baku start")

	var (
		parseFile = flag.String("c", "etc/baku/baku.toml", "specify config file")
	)

	flag.Parse()

	conf := LoadConfig(*parseFile)

	database := LoadDatabase(conf.Database, "data")
	database.Setup(conf.Job.Group)

	scheduler := InitScheduler()
	scheduler.Run(conf.Job.Group, database)

	server := InitServer(conf.Server)
	server.Run(conf.Job, database)
}
