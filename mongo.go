package log

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoLogger struct {
	Logger
}

func NewMongoLogger(logger Logger) *MongoLogger {
	return &MongoLogger{Logger: logger}
}

func (l MongoLogger) Options() *options.LoggerOptions {
	return options.
		Logger().
		SetSink(l).
		SetMaxDocumentLength(10240).
		SetComponentLevel(options.LogComponentCommand, options.LogLevelDebug)
}
func (l MongoLogger) Info(level int, msg string, data ...interface{}) {
	switch options.LogLevel(level) {
	case options.LogLevelDebug:
		l.Logger.Debug(msg, anyToZapFieldMongo(data...)...)
	case options.LogLevelInfo:
		l.Logger.Info(msg, anyToZapFieldMongo(data...)...)
	default:
		l.Logger.Info(msg, anyToZapFieldMongo(data...)...)
	}
}
func (l MongoLogger) Error(err error, msg string, data ...interface{}) {
	l.Logger.Error(msg, append(anyToZapFieldMongo(data...), zap.Error(err))...)
}

func anyToZapFieldMongo(data ...any) []zap.Field {
	var (
		result = make([]zap.Field, 0, len(data))
	)

	for i := 0; i < len(data); i += 2 {
		if i+1 < len(data) {
			result = append(result, zap.Any(data[i].(string), data[i+1]))
		} else {
			result = append(result, zap.Any(fmt.Sprintf(`数据%d`, i+1), data[i]))
		}
	}

	return result
}

func (l MongoLogger) CommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			l.Logger.Info(`开始执行`, zap.Int64(`请求ID`, startedEvent.RequestID), zap.Any(`command`, startedEvent.Command.String()))
		},
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
			id := succeededEvent.RequestID
			l.Logger.Info(`执行成功`, zap.Int64(`请求ID`, id), zap.Duration(`耗时`, succeededEvent.Duration), zap.Any(`result`, succeededEvent.Reply.String()))
		},
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
			id := failedEvent.RequestID
			l.Logger.Info(`执行失败`, zap.Int64(`请求ID`, id), zap.Duration(`耗时`, failedEvent.Duration), zap.String(`原因`, failedEvent.Failure))
		},
	}
}
