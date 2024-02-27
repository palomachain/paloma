package liblog_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/stretchr/testify/assert"
)

type mock struct {
	invocations []string
	args        []interface{}
}

func (m *mock) Impl() interface{} {
	return m
}

func (m *mock) Debug(msg string, keyvals ...interface{}) {
	m.invocations = append(m.invocations, "Debug")
	m.args = append(m.args, msg)
	m.args = append(m.args, keyvals...)
}

func (m *mock) Error(msg string, keyvals ...interface{}) {
	m.invocations = append(m.invocations, "Error")
	m.args = append(m.args, msg)
	m.args = append(m.args, keyvals...)
}

func (m *mock) Info(msg string, keyvals ...interface{}) {
	m.invocations = append(m.invocations, "Info")
	m.args = append(m.args, msg)
	m.args = append(m.args, keyvals...)
}

func (m *mock) With(keyvals ...interface{}) log.Logger {
	m.invocations = append(m.invocations, "With")
	m.args = append(m.args, keyvals...)
	return m
}

func (m *mock) Warn(msg string, keyvals ...any) {
	m.invocations = append(m.invocations, "Warn")
	m.args = append(m.args, msg)
	m.args = append(m.args, keyvals...)
}

func Test_Liblog(t *testing.T) {
	testErr := fmt.Errorf("Romanes eunt domus")
	for _, tt := range []struct {
		run  func(l liblog.Logr)
		name string
		inv  []string
		args []interface{}
	}{
		{
			name: "Debug",
			inv:  []string{"Debug"},
			args: []interface{}{"log message", "msg-id", 2},
			run: func(l liblog.Logr) {
				l.Debug("log message", "msg-id", 2)
			},
		},
		{
			name: "Info",
			inv:  []string{"Info"},
			args: []interface{}{"log message", "msg-id", 2},
			run: func(l liblog.Logr) {
				l.Info("log message", "msg-id", 2)
			},
		},
		{
			name: "Error",
			inv:  []string{"Error"},
			args: []interface{}{"log message", "msg-id", 2},
			run: func(l liblog.Logr) {
				l.Error("log message", "msg-id", 2)
			},
		},
		{
			name: "With",
			inv:  []string{"With"},
			args: []interface{}{"extra-data", 1510},
			run: func(l liblog.Logr) {
				l.With("extra-data", 1510)
			},
		},
		{
			name: "WithFields",
			inv:  []string{"With"},
			args: []interface{}{"extra-data", 1510},
			run: func(l liblog.Logr) {
				l.WithFields("extra-data", 1510)
			},
		},
		{
			name: "WithError",
			inv:  []string{"With"},
			args: []interface{}{"error", testErr},
			run: func(l liblog.Logr) {
				l.WithError(testErr)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := mock{
				invocations: []string{},
				args:        []interface{}{},
			}
			l := liblog.FromSDKLogger(&m)
			tt.run(l)

			assert.Equal(t, tt.inv, m.invocations, "unexpected invocations, want: %v, got: %v", tt.inv, m.invocations)
			assert.Equal(t, tt.args, m.args, "unexpected args, want: %v, got: %v", tt.args, m.args)
		})
	}
}
