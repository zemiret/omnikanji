package main

import (
	"log"
)

type Logger struct {
	log.Logger
}

func (lo *Logger) Errorf(err error, format string, v ...interface{}) {
	lo.Printf(format+": "+err.Error()+"\n", v)
}
