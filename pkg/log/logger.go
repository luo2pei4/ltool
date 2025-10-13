package logger

import (
	"log"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *logrus.Logger

func InitLog(logLevel string, logPath string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Printf("%s, %s", err.Error(), "log level must be info/warn/error/debug")
		level = logrus.InfoLevel
	}
	logger = logrus.New()
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(&lumberjack.Logger{
		Filename: logPath,
		MaxSize:  10,
		MaxAge:   10,
		Compress: true,
	})
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Info(args ...string) {
	logger.Info(args)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Warn(args ...string) {
	logger.Warn(args)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Error(args ...string) {
	logger.Error(args)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Debug(args ...string) {
	logger.Debug(args)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func SetLogLevel(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Printf("%s, %s", err.Error(), "log level must be info/warn/error/debug")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
}
