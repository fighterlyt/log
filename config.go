package log

import (
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

const (
	// defaultTimeZone 默认时区
	defaultTimeZone = `Asia/Shanghai`
	// defaultTimeLayout 默认时间格式
	defaultTimeLayout = `2006-01-02 15:04:05.000`
)

var (
	core    zapcore.Core
	encoder zapcore.Encoder
	w       zapcore.WriteSyncer
)

// Config 日志器配置
type Config struct {
	Service    string         `yaml:"service"`    // 日志名称
	Level      zapcore.Level  `yaml:"level"`      // 最低级别
	FilePath   string         `yaml:"filePath"`   // 日志文件路径,如果为空，表示不输出，可以包含路径,最终生成一个FilePath.log.
	TimeZone   string         `yaml:"timeZone"`   // 时区，默认defaultTimeZone,可以从https://www.zeitverschiebung.net/en/ 查询时区信息
	TimeLayout string         `yaml:"timeLayout"` // 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
	Debug      bool           `yaml:"debug"`      // 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的代码行信息
	location   *time.Location `yaml:"location"`
}

func NewConfig() *Config {
	return &Config{}
}

func NewConfigFromYamlData(yamlData io.Reader) (config *Config, err error) {
	config = NewConfig()
	if err = yaml.NewDecoder(yamlData).Decode(config); err != nil {
		return nil, errors.Wrap(err, `解析错误`)
	}

	return config, nil
}

func (l *Config) Build() (logger Logger, err error) {
	var (
		underlyingLogger *zap.Logger
		allCores         []zapcore.Core
	)

	cfg := &zap.Config{
		Level:            zap.NewAtomicLevelAt(l.Level),
		Development:      true,
		Encoding:         "console",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if l.TimeZone == `` {
		l.TimeZone = defaultTimeZone
	}

	if l.TimeLayout == `` {
		l.TimeLayout = defaultTimeLayout
	}

	if l.location, err = time.LoadLocation(l.TimeZone); err != nil {
		return nil, errors.Wrapf(err, `加载时区[%s]`, l.TimeZone)
	}

	// todo: 如何验证一个time layout 是否正确

	cfg.EncoderConfig = l.NewDevelopmentEncoderConfig()

	encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)

	if l.FilePath != `` {
		w = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&lumberjack.Logger{
			Filename:   l.FilePath + ".log",
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		}))

		allCores = append(allCores, zapcore.NewCore(
			encoder,
			w,
			cfg.Level,
		), zapcore.NewCore(encoder, os.Stdout, cfg.Level))
	} else {
		allCores = append(allCores, zapcore.NewCore(encoder, os.Stdout, cfg.Level))
	}

	core = zapcore.NewTee(allCores...)
	underlyingLogger = zap.New(core, zap.AddCaller())

	return newLogger(underlyingLogger.With(zap.String(`系统`, l.Service)), ``, 1, true), nil
}

func (l *Config) NewDevelopmentEncoderConfig() zapcore.EncoderConfig {
	config := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:       "T",
		LevelKey:      "L",
		NameKey:       "N",
		CallerKey:     "C",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "M",
		StacktraceKey: "S",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(l.location).Format(l.TimeLayout))
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	if !l.Debug {
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return config
}
