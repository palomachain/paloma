package liblog

import (
	"context"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Logr interface {
	log.Logger
	WithFields(keyvals ...interface{}) Logr
	WithError(err error) Logr
	WithComponent(string) Logr
	WithValidator(string) Logr
}

type LogProvider interface {
	Logger(sdk.Context) log.Logger
}

func FromKeeper(ctx context.Context, p LogProvider) Logr {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return FromSDKLogger(p.Logger(sdkCtx))
}

func FromSDKLogger(l log.Logger) Logr {
	return &lgwr{l}
}

type lgwr struct {
	l log.Logger
}

// Impl implements Logr.
func (*lgwr) Impl() any {
	return log.NewNopLogger()
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

func (l *lgwr) Warn(msg string, keyvals ...any) {
	l.l.Warn(msg)
}
