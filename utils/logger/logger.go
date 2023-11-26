package logger

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

var L *logrus.Logger

const DateTimeFormat = "2006-01-02 15:04:05"

func init() {
	//Инициализируем логер
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
