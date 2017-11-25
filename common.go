package main

import (
	"encoding/json"

	"go.uber.org/zap"
)

func Contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func Logger(level string, tagName string, message string) {
	raw := []byte(`{
	  "level": "debug",
	  "encoding": "console",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
			"levelKey": "level",
			"timeKey": "Time",
			"nameKey": "Name",
			"levelEncoder": "lowercase",
			"timeEncoder": "iso8601"
	  }
	}`)

	var zapConfig zap.Config
	if err := json.Unmarshal(raw, &zapConfig); err != nil {
		panic(err)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	logger = logger.Named(tagName)

	switch level {
	case "info":
		logger.Info(message)
	case "warn":
		logger.Warn(message)
	case "error":
		logger.Error(message)
	case "fatal":
		logger.Fatal(message)
	default:
		panic("level missing")
	}
}
