package main

import (
	"math/rand"
	"os"

	"github.com/sirupsen/logrus"
)

// LogConfig configures the cxMate logger.
type LogConfig struct {
	Debug  bool
	File   string
	Format string
}

// Logger has methods for logging cxMate events and errors.
type Logger struct {
	log *logrus.Entry
}

// NewLogger creates a new cxMate service logger using a logger configuration.
func (c LogConfig) NewLogger(service string, version string) (*Logger, error) {
	var formatter logrus.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	if c.Format == "json" {
		formatter = &logrus.JSONFormatter{}
	}
	out := os.Stderr
	if c.File != "" {
		f, err := os.OpenFile(c.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0666)
		if err != nil {
			return nil, err
		}
		out = f
	}
	level := logrus.InfoLevel
	if c.Debug {
		level = logrus.DebugLevel
	}
	l := &logrus.Logger{
		Out:       out,
		Formatter: formatter,
		Level:     level,
	}
	logger := &Logger{
		log: l.WithFields(logrus.Fields{
			"id":      randID(10),
			"service": service,
			"version": version,
		}),
	}
	return logger, nil
}

func randID(n int) string {
	b := make([]rune, n)

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//AddField reates a new logger with a new field added to entries.
func (l *Logger) AddField(field string, value interface{}) *Logger {
	return &Logger{log: l.log.WithField(field, value)}
}

//AddFields reates a new logger with new fields added to entries.
func (l *Logger) AddFields(fields map[string]interface{}) *Logger {
	return &Logger{log: l.log.WithFields(fields)}
}

// Info logs a message at level Info on the logger.
func (l *Logger) Info(args ...interface{}) {
	l.log.Info(args)
}

// Infoln logs a message at level Info on the logger.
func (l *Logger) Infoln(args ...interface{}) {
	l.log.Infoln(args)
}

// Debug logs a message at level Debug on the logger.
func (l *Logger) Debug(args ...interface{}) {
	l.log.Debug(args)
}

// Debugln logs a message at level Debug on the logger.
func (l *Logger) Debugln(args ...interface{}) {
	l.log.Debugln(args)
}

// Error logs a message at level Error on the logger.
func (l *Logger) Error(args ...interface{}) {
	l.log.Error(args)
}

// Errorln logs a message at level Error on the logger.
func (l *Logger) Errorln(args ...interface{}) {
	l.log.Errorln(args)
}

// Fatal logs a message at level Fatal on the logger.
func (l *Logger) Fatal(args ...interface{}) {
	l.log.Fatal(args)
}

// Fatalln logs a message at level Fatal on the logger.
func (l *Logger) Fatalln(args ...interface{}) {
	l.log.Fatalln(args)
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
