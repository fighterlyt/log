package log

import (
	"io"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	debug = false

	debugPrintln = func(args ...interface{}) {
		if debug {
			log.Println(args...)
		}
	}
)

type Hook interface {
	Writer() io.Writer
	MinLevel() zapcore.Level
}

// Logger 日志器接口
type Logger interface {
	// Derive 衍生新的日志器,name会进入名称，多个名称会成为name.name
	Derive(name string) Logger
	// With 添加某些字段
	With(fields ...zap.Field) Logger
	WithWhenNotExist(key string, field zap.Field) Logger
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
	// Start 返回一个携带任务ID字段的日志器
	Start() Logger
	// SetLevel 设置级别，可以调高或者调低
	SetLevel(level zapcore.Level) Logger
	AddCallerSkip(skip int) Logger
}

func ensureDuplicateKeys(data *Exist) *Exist {
	if data != nil {
		return data
	}

	return NewExist(10)
}

// logger 日志器的实现
type logger struct {
	underlying    *zap.Logger
	levelToPath   map[zapcore.Level]string
	duplicateKeys *Exist
	name          string
	fields        []zapcore.Field
	skip          int
}

/*
newLogger 生成一个日志器
参数:
*	underlying	*zap.Logger     			底层日志器
*	name      	string          			对应的名称
*	skip      	int             			跳过的堆栈
*	setName   	bool            			是否需要设置名称
*	last      	bool            			是否名称只需要最后一段
* 	levelToPath map[zapcore.Level]string	不同级别重定向
*	fields    	...zapcore.Field			字段
返回值:
*	*logger   	*logger         	日志器
*/
func newLogger(underlying *zap.Logger, name string, skip int, setName, last bool, levelToPath map[zapcore.Level]string, duplicateKeys *Exist, fields ...zapcore.Field) *logger { //nolint:lll
	result := &logger{
		underlying:    underlying,
		name:          name,
		duplicateKeys: ensureDuplicateKeys(duplicateKeys),
		fields:        fields,
	}

	debugPrintln(`newLogger`, name, setName, skip)

	if setName {
		if last {
			nameFields := strings.Split(name, `.`)
			debugPrintln(`named`, nameFields[len(nameFields)-1])
			result.underlying = result.underlying.Named(nameFields[len(nameFields)-1])
		} else {
			result.underlying = result.underlying.Named(name)
		}
	}

	if skip >= 0 {
		result.underlying = result.underlying.WithOptions(zap.AddCallerSkip(skip))
		result.skip = skip
	}

	result.levelToPath = levelToPath

	return result
}

/*
Derive 衍生出一个新的子日志器
参数:
*	s     	string	名称
返回值:
*	Logger	Logger	日志器
*/
func (l *logger) Derive(s string) Logger {
	debugPrintln(`derive`, s)

	var names []string
	if l.name == `` {
		names = append(names, s)
	} else {
		names = append(names, l.name, s)
	}

	return newLogger(l.underlying, strings.Join(names, "."), -1, true, true, l.levelToPath, l.duplicateKeys.Copy(), l.fields...)
}

func (l logger) With(fields ...zap.Field) Logger {
	fields = append(l.fields, fields...)

	return newLogger(l.underlying.With(fields...), l.name, -1, false, false, l.levelToPath, l.duplicateKeys.Copy())
}

func (l logger) WithWhenNotExist(key string, field zap.Field) Logger {
	// 判断是否存在，存在就返回l
	if l.duplicateKeys.Exist(key) {
		return &l
	}

	fields := make([]zap.Field, len(l.fields))
	copy(fields, l.fields)

	// 不存在就直接调用
	fields = append(fields, field)

	duplicate := l.duplicateKeys.Copy()

	duplicate.Set(key)

	return newLogger(l.underlying.With(fields...), l.name, -1, false, false, l.levelToPath, duplicate)
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
	debugPrintln(`setLevel`, level, l.name)

	var allCore []zapcore.Core
	if writeSyncer != nil {
		allCore = append(allCore, zapcore.NewCore(
			encoder,
			writeSyncer,
			level,
		))
	}

	for _, inputCore := range inputCores {
		if inputCore != nil {
			allCore = append(allCore, inputCore)
		}
	}

	if !HiddenConsole {
		allCore = append(allCore, zapcore.NewCore(encoder, os.Stdout, level))
	}

	core = zapcore.NewTee(allCore...)

	resultLogger := zap.New(core).With(l.fields...)
	resultLogger = resultLogger.WithOptions(zap.AddCaller())

	result := newLogger(resultLogger, l.name, 1, true, false, l.levelToPath, nil, l.fields...)

	return result
}

func (l *logger) Start() Logger {
	return l.With(zap.String(`任务ID`, primitive.NewObjectID().Hex()))
}

func (l *logger) AddCallerSkip(skip int) Logger {
	return newLogger(l.underlying, l.name, skip, false, false, l.levelToPath, l.duplicateKeys)
}
