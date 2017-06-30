package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

//LogConfig configures the cxMate logger
type LogConfig struct {
	Debug  bool
	File   string
	Format string
}

func configureLogger(c LogConfig) error {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if c.File != "" {
		f, err := os.OpenFile(c.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0666)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
	}
	switch c.Format {
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	}
	return nil
}

func logDebug(args ...interface{}) {
	logrus.Debug(args)
}

func logDebugln(args ...interface{}) {
	logrus.Debugln(args)
}

func logFatal(args ...interface{}) {
	logrus.Fatal(args)
}

func logFatalln(args ...interface{}) {
	logrus.Fatalln(args)
}
