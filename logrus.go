package log

import (
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar/log"
	"go.uber.org/zap"
)

type pulsarLogger struct {
	Logger
}

func NewPulsarLogger(logger Logger) *pulsarLogger {
	return &pulsarLogger{Logger: logger}
}

func (p pulsarLogger) SubLogger(fields log.Fields) log.Logger {
	zapFields := make([]zap.Field, 0, len(fields))

	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return NewPulsarLogger(p.Logger.With(zapFields...))
}

func (p pulsarLogger) WithFields(fields log.Fields) log.Entry {
	return p.SubLogger(fields)
}

func (p pulsarLogger) WithField(name string, value interface{}) log.Entry {
	return NewPulsarLogger(p.Logger.With(zap.Any(name, value)))
}

func (p pulsarLogger) WithError(err error) log.Entry {
	return NewPulsarLogger(p.Logger.With(zap.Error(err)))
}

func (p pulsarLogger) Debug(args ...interface{}) {
	p.Logger.Debug(``, zap.Any(`参数`, args))
}

func (p pulsarLogger) Info(args ...interface{}) {
	p.Logger.Info(``, zap.Any(`参数`, args))
}

func (p pulsarLogger) Warn(args ...interface{}) {
	p.Logger.Warn(``, zap.Any(`参数`, args))
}

func (p pulsarLogger) Error(args ...interface{}) {
	p.Logger.Error(``, zap.Any(`参数`, args))
}

func (p pulsarLogger) Debugf(format string, args ...interface{}) {
	p.Logger.Debug(fmt.Sprintf(format, args...))
}

func (p pulsarLogger) Infof(format string, args ...interface{}) {
	p.Logger.Info(fmt.Sprintf(format, args...))
}

func (p pulsarLogger) Warnf(format string, args ...interface{}) {
	p.Logger.Warn(fmt.Sprintf(format, args...))
}

func (p pulsarLogger) Errorf(format string, args ...interface{}) {
	p.Logger.Error(fmt.Sprintf(format, args...))
}
