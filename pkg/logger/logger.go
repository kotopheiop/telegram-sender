package logger

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

var L *logrus.Logger

const DateTimeFormat = "2006-01-02 15:04:05 MST"

// Инициализируем логер
func init() {
	L = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &prefixed.TextFormatter{
			TimestampFormat: DateTimeFormat,
			FullTimestamp:   true,
			ForceFormatting: true,
			ForceColors:     true,
		},
	}
}

// Объявим обёртки, чтобы не вызывть logger.L.Fatal...

func Fatal(args ...interface{}) {
	L.Fatal(args...)
}

func Info(args ...interface{}) {
	L.Info(args...)
}

func Errorf(format string, args ...interface{}) {
	L.Errorf(format, args...)
}

func Infof(format string, args ...interface{}) {
	L.Infof(format, args...)
}

func Warningf(format string, args ...interface{}) {
	L.Warningf(format, args...)
}
