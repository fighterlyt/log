package log

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type cronLogger struct {
	Logger
}

func (l cronLogger) Printf(format string, values ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, values...))
}

func (l cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Logger.Error(fmt.Sprintf(msg, keysAndValues...), zap.Error(err))
}

func NewCronLogger(targetLogger Logger) *cronLogger {
	return &cronLogger{Logger: targetLogger}
}

func DeriveCronLogger(baseLogger Logger, topic, method string) Logger {
	return baseLogger.With(zap.Strings(`topic/method`, []string{topic, method}))
}

// cronFormatTimes formats any time.Time values as RFC3339. 这块来自cron库
func cronFormatTimes(keysAndValues []interface{}) []interface{} {
	var formattedArgs []interface{}

	for _, arg := range keysAndValues {
		if t, ok := arg.(time.Time); ok {
			arg = t.Format(time.RFC3339)
		}

		formattedArgs = append(formattedArgs, arg)
	}

	return formattedArgs
}

// cronFormatString returns a logfmt-like format string for the number of
// key/values. 注意:来自cron库
func cronFormatString(numKeysAndValues int) string {
	var sb strings.Builder

	sb.WriteString("%s")

	if numKeysAndValues > 0 {
		sb.WriteString(", ")
	}

	for i := 0; i < numKeysAndValues/2; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("%v=%v")
	}

	return sb.String()
}
