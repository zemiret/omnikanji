package logger 

import (
	"log"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.Default(),
	}
}

func (lo *Logger) Errorf(err error, format string, v ...interface{}) {
	lo.Printf(format+": "+err.Error()+"\n", v)
}
