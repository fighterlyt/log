package log

import (
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志器接口
type Logger interface {
	// Derive 衍生新的日志器,name会进入名称，多个名称会成为name.name
	Derive(name string) Logger
	// With 添加某些字段
	With(fields ...zap.Field) Logger
	// Debug 输出日志到Debug 级别
	Debug(msg string, fields ...zap.Field)
	// Info 输出日志到Info 级别
	Info(msg string, fields ...zap.Field)
	// Warn 输出日志到Warn 级别
	Warn(msg string, fields ...zap.Field)
	// Error 输出日志到Error 级别
	Error(msg string, fields ...zap.Field)
	// Fatal 输出日志到Fatal 级别
	Fatal(msg string, fields ...zap.Field)
	// Panic 输出日志到Panic 级别
	Panic(msg string, fields ...zap.Field)
	// Start() 返回一个携带任务ID字段的日志器
	Start() Logger
	// SetLevel 设置级别，可以调高或者调低
	SetLevel(level zapcore.Level) Logger
	AddCallerSkip(skip int) Logger
}

// logger 日志器的实现
type logger struct {
	underlying *zap.Logger     // 底层日志器
	fields     []zapcore.Field // 嵌套的字段
	name       string          // 对应的名称
}

/*newLogger 生成一个日志器
参数:
*	underlying	*zap.Logger         底层日志器
*	name      	string              对应的名称
*	fields    	...zapcore.Field    字段
返回值:
*	*logger	*logger
*/
func newLogger(underlying *zap.Logger, name string, skip int, setName bool, fields ...zapcore.Field) *logger {
	result := &logger{underlying: underlying, name: name, fields: fields}

	if setName {
		result.underlying = result.underlying.Named(name)
	}

	if skip >= 0 {
		result.underlying = result.underlying.WithOptions(zap.AddCallerSkip(skip))
	}

	return result
}

func (l *logger) Derive(s string) Logger {
	var names []string
	if l.name == `` {
		names = append(names, s)
	} else {
		names = append(names, l.name, s)
	}

	return newLogger(l.underlying, strings.Join(names, "."), -1, true, l.fields...)
}

func (l logger) With(fields ...zap.Field) Logger {
	fields = append(l.fields, fields...)
	return newLogger(l.underlying.With(fields...), l.name, -1, false, fields...)
}

func (l logger) Info(msg string, fields ...zap.Field) {
	l.underlying.Info(msg, fields...)
}

func (l logger) Debug(msg string, fields ...zap.Field) {
	l.underlying.Debug(msg, fields...)
}

func (l logger) Warn(msg string, fields ...zap.Field) {
	l.underlying.Warn(msg, fields...)
}

func (l logger) Error(msg string, fields ...zap.Field) {
	l.underlying.Error(msg, fields...)
}

func (l logger) Fatal(msg string, fields ...zap.Field) {
	l.underlying.Fatal(msg, fields...)
}

func (l logger) Panic(msg string, fields ...zap.Field) {
	l.underlying.Panic(msg, fields...)
}

func (l logger) SetLevel(level zapcore.Level) Logger {
	var allCore []zapcore.Core
	if w != nil {
		allCore = append(allCore, zapcore.NewCore(
			encoder,
			w,
			level,
		))
	}

	allCore = append(allCore, zapcore.NewCore(encoder, os.Stdout, level))

	core = zapcore.NewTee(allCore...)

	logger := zap.New(core).With(l.fields...)
	logger = logger.WithOptions(zap.AddCaller())

	return newLogger(logger, l.name, -1, true, l.fields...)
}

func (l *logger) Start() Logger {
	return l.With(zap.String(`任务ID`, primitive.NewObjectID().Hex()))
}

func (l *logger) AddCallerSkip(skip int) Logger {
	return newLogger(l.underlying, l.name, skip, false)
}
