// Package log contain logger interface and implementations
package log

import (
	"runtime"
	"strings"
)

const LoggerComponentKey = "instance"

type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})

	With(args ...interface{}) Logger

	Debug(args ...interface{})
	Debugf(template string, args ...interface{})

	Info(args ...interface{})
	Infof(template string, args ...interface{})

	Warn(args ...interface{})
	Warnf(template string, args ...interface{})

	Error(args ...interface{})
	Errorf(template string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
}

func ForkLogger(logger Logger, args ...interface{}) Logger {
	if len(args) == 0 {
		var p string
		pc, _, _, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		if ok && details != nil {
			p = details.Name()
		}

		parts := strings.Split(p, "New")
		if len(parts) > 1 {
			return logger.With(LoggerComponentKey, parts[1])
		}
		return logger.With(LoggerComponentKey, p)
	}
	return logger.With(LoggerComponentKey, args[0])
}
