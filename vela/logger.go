package vela

import "go.uber.org/zap/zapcore"

type LogByEnv interface {
	Log
	WithLogger(Log)
	WithLevel(zapcore.Level)
	LoggerLevel() zapcore.Level
}

type Log interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Panic(...interface{})
	Fatal(...interface{})

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})
}
