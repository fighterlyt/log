package log

import (
	"fmt"

	microlog "go-micro.dev/v4/logger"
	"go.uber.org/zap"
)

type microLogger struct {
	Logger
	options *microlog.Options
}

func (m *microLogger) Init(options ...microlog.Option) error {
	for _, option := range options {
		option(m.options)
	}

	return nil
}

func (m microLogger) Options() microlog.Options {
	return *m.options
}

func (m microLogger) Fields(fields map[string]interface{}) microlog.Logger {
	zapFields := make([]zap.Field, 0, len(fields))

	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	if m.options.CallerSkipCount >= 0 {
		m.Logger = m.Logger.AddCallerSkip(m.options.CallerSkipCount)
	}

	return NewMicroLogger(m.Logger.With(zapFields...))
}

func (m microLogger) Log(level microlog.Level, v ...interface{}) {
	switch level {
	case microlog.InfoLevel:
		m.Logger.Info(``, zap.Any(``, v))
	case microlog.DebugLevel, microlog.TraceLevel:
		m.Logger.Debug(``, zap.Any(``, v))
	case microlog.WarnLevel:
		m.Logger.Warn(``, zap.Any(``, v))
	case microlog.ErrorLevel:
		m.Logger.Error(``, zap.Any(``, v))
	case microlog.FatalLevel:
		m.Logger.Fatal(``, zap.Any(``, v))
	default:
		m.Logger.Info(``, zap.Any(``, v))
	}
}

func (m microLogger) Logf(level microlog.Level, format string, v ...interface{}) {
	switch level {
	case microlog.InfoLevel:
		m.Logger.Info(fmt.Sprintf(format, v...))
	case microlog.DebugLevel, microlog.TraceLevel:
		m.Logger.Debug(fmt.Sprintf(format, v...))
	case microlog.WarnLevel:
		m.Logger.Warn(fmt.Sprintf(format, v...))
	case microlog.ErrorLevel:
		m.Logger.Error(fmt.Sprintf(format, v...))
	case microlog.FatalLevel:
		m.Logger.Fatal(fmt.Sprintf(format, v...))
	default:
		m.Logger.Info(fmt.Sprintf(format, v...))
	}
}

func (m microLogger) String() string {
	return `zap-micro`
}

func NewMicroLogger(logger Logger) *microLogger {
	return &microLogger{
		Logger:  logger,
		options: &microlog.Options{},
	}
}
