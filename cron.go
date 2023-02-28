package log

import (
	"fmt"

	"go.uber.org/zap"
)

type cronLogger struct {
	Logger
}

func (l cronLogger) Printf(format string, values ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, values...))
}

func NewCronLogger(targetLogger Logger) *cronLogger {
	return &cronLogger{Logger: targetLogger}
}

func DeriveCronLogger(baseLogger Logger, topic, method string) Logger {
	return baseLogger.With(zap.Strings(`topic/method`, []string{topic, method}))
}
