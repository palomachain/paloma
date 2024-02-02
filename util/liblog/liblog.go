package liblog

import (
	"github.com/cometbft/cometbft/libs/log"
)

type Logr interface {
	log.Logger
	WithFields(keyvals ...interface{}) Logr
	WithError(err error) Logr
	WithComponent(string) Logr
	WithValidator(string) Logr
}

func FromSDKLogger(l log.Logger) Logr {
	return &lgwr{l: l}
}

type lgwr struct {
	l log.Logger
}

func (l *lgwr) Debug(msg string, keyvals ...interface{}) {
	l.l.Debug(msg, keyvals...)
}

func (l *lgwr) Error(msg string, keyvals ...interface{}) {
	l.l.Error(msg, keyvals...)
}

func (l *lgwr) Info(msg string, keyvals ...interface{}) {
	l.l.Info(msg, keyvals...)
}

func (l *lgwr) With(keyvals ...interface{}) log.Logger {
	return l.l.With(keyvals...)
}

func (l *lgwr) WithFields(keyvals ...interface{}) Logr {
	l.l = l.l.With(keyvals...)
	return l
}

func (l *lgwr) WithComponent(c string) Logr {
	l.l = l.l.With("component", c)
	return l
}

func (l *lgwr) WithValidator(v string) Logr {
	l.l = l.l.With("validator", v)
	return l
}

func (l *lgwr) WithError(err error) Logr {
	return l.WithFields("error", err)
}
