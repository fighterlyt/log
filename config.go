package log

import (
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

const (
	// defaultTimeZone 默认时区
	defaultTimeZone = `Asia/Shanghai`
	// defaultTimeLayout 默认时间格式
	defaultTimeLayout       = `2006-01-02 15:04:05.000`
	defaultRotateMaxSize    = 100
	defaultRotateMaxBackups = 50
	defaultRotateMaxAge     = 7
)

var (
	core          zapcore.Core
	encoder       zapcore.Encoder
	writeSyncer   zapcore.WriteSyncer
	inputCores    []zapcore.Core
	HiddenConsole bool
)

// RotateConfig rotate 配置
type RotateConfig struct {
	MaxSize         int  `yaml:"maxSize"`         // 单个日志文件最大大小，单位为MB
	MaxBackups      int  `yaml:"maxBackups"`      // 最大部分数量
	MaxAge          int  `yaml:"maxAge"`          // 最大保留时间,单位为天
	DisableCompress bool `yaml:"disableCompress"` // 不压缩
}

// Config 日志器配置
type Config struct {
	Service     string            `yaml:"service"`     // 日志名称
	Level       zapcore.Level     `yaml:"level"`       // 最低级别
	FilePath    string            `yaml:"filePath"`    // 日志文件路径,如果为空，表示不输出，可以包含路径,最终生成一个FilePath.log.
	TimeZone    string            `yaml:"timeZone"`    // 时区，默认defaultTimeZone,可以从https://www.zeitverschiebung.net/en/ 查询时区信息
	TimeLayout  string            `yaml:"timeLayout"`  // 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
	Debug       bool              `yaml:"debug"`       // 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的代码行信息
	Dev         bool              `yaml:"dev"`         // 开发模式，输出完整路径
	JSON        bool              `yaml:"json"`        // 是否输出为一个完整的json,默认为false
	HideConsole bool              `yaml:"hideConsole"` // 是否隐藏终端输出
	Rotate      *RotateConfig     `yaml:"rotate"`      // 日志 rotate
	location    *time.Location    `yaml:"location"`    // 输出time.Time时的时区
	LevelToPath map[string]string `yaml:"levelToPath"`
	levelToPath map[zapcore.Level]string
	Hooks       []Hook
}

/*
NewConfig 新建一个配置
参数:
返回值:
*	*Config	*Config
*/
func NewConfig() *Config {
	return &Config{}
}

func (l *Config) tidy() error {
	var (
		level zapcore.Level
		err   error
	)

	l.levelToPath = make(map[zapcore.Level]string, len(l.LevelToPath))

	for levelText, path := range l.LevelToPath {
		if level, err = zapcore.ParseLevel(levelText); err != nil {
			return errors.Wrapf(err, `解析level[%s]`, levelText)
		}
		l.levelToPath[level] = path
	}

	return nil
}

/*
NewConfigFromYamlData 从yaml数据中新建配置
参数:
*	yamlData	io.Reader   yaml数据 reader，不能为空
返回值:
*	config	*Config
*	err   	error
*/
func NewConfigFromYamlData(yamlData io.Reader) (config *Config, err error) {
	config = NewConfig()
	if err = yaml.NewDecoder(yamlData).Decode(config); err != nil {
		return nil, errors.Wrap(err, `解析错误`)
	}

	return config, nil
}

/*
NewConfigFromToml 从toml配置中构建
参数:
*	tomlData	[]byte 	参数1
返回值:
*	config  	*Config	返回值1
*	err     	error  	返回值2
*/
func NewConfigFromToml(tomlData []byte) (config *Config, err error) {
	config = NewConfig()
	if err = toml.Unmarshal(tomlData, config); err != nil {
		return nil, errors.Wrap(err, `解析错误`)
	}

	return config, nil
}

/*
Build 构建日志器
参数:
返回值:
*	logger	Logger  日志器
*	err   	error   错误
*/
func (l *Config) Build(cores ...zapcore.Core) (logger Logger, err error) {
	var (
		underlyingLogger *zap.Logger
		allCores         []zapcore.Core
	)

	if err = l.tidy(); err != nil {
		return nil, errors.Wrap(err, `tidy`)
	}

	HiddenConsole = l.HideConsole
	inputCores = cores

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

	cfg.EncoderConfig = l.newEncoderConfig()

	if l.JSON {
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	}

	if l.FilePath != `` {
		if l.Rotate == nil {
			l.Rotate = &RotateConfig{}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   l.FilePath + ".log",
			MaxSize:    l.Rotate.MaxSize, // megabytes
			MaxBackups: l.Rotate.MaxBackups,
			MaxAge:     l.Rotate.MaxAge, // days
			Compress:   !l.Rotate.DisableCompress,
		}

		fillLumberjack(lumberjackLogger)

		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberjackLogger))

		allCores = append(allCores, zapcore.NewCore(
			encoder,
			writeSyncer,
			newLevelEnablerWithExcept(cfg.Level, l.levelToPath),
		))
	}

	if l.levelToPath != nil {
		for level := range l.levelToPath {
			lumberjackLogger := &lumberjack.Logger{
				Filename:   l.levelToPath[level],
				MaxSize:    l.Rotate.MaxSize, // megabytes
				MaxBackups: l.Rotate.MaxBackups,
				MaxAge:     l.Rotate.MaxAge, // days
				Compress:   true,
			}

			fillLumberjack(lumberjackLogger)

			writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberjackLogger))

			allCores = append(allCores, zapcore.NewCore(encoder, writeSyncer, newLevelEnablerWithExcept(level, l.levelToPath, level)))
		}
	}

	if !l.HideConsole {
		allCores = append(allCores, zapcore.NewCore(encoder, os.Stdout, cfg.Level))
	}

	for i := range l.Hooks {
		hook := l.Hooks[i]

		allCores = append(allCores, zapcore.NewCore(encoder, zapcore.AddSync(hook.Writer()), hook.MinLevel()))
	}

	allCores = append(allCores, cores...)

	core = zapcore.NewTee(allCores...)
	underlyingLogger = zap.New(core, zap.AddCaller())

	return newLogger(underlyingLogger.With(zap.String(`系统`, l.Service)), ``, 1, true, false, l.levelToPath), nil
}

func NewEasyLogger(debug, hideConsole bool, filePath, service string) (Logger, error) {
	config := NewConfig()
	config.Debug = debug
	config.FilePath = filePath
	config.HideConsole = hideConsole
	config.Service = service

	return config.Build()
}

/*
newEncoderConfig 新建编码器配置
参数:
返回值:
*	zapcore.EncoderConfig	zapcore.EncoderConfig
*/
func (l *Config) newEncoderConfig() zapcore.EncoderConfig {
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
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(l.location).Format(l.TimeLayout))
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if l.Dev {
		config.EncodeCaller = zapcore.FullCallerEncoder
	}

	return config
}

func fillLumberjack(lumberjackLogger *lumberjack.Logger) {
	if lumberjackLogger.MaxSize == 0 {
		lumberjackLogger.MaxSize = defaultRotateMaxSize
	}

	if lumberjackLogger.MaxAge == 0 {
		lumberjackLogger.MaxAge = defaultRotateMaxAge
	}

	if lumberjackLogger.MaxBackups == 0 {
		lumberjackLogger.MaxBackups = defaultRotateMaxBackups
	}
}

type levelEnableWithExcept struct {
	zapcore.LevelEnabler
	except map[zapcore.Level]bool
}

func (l levelEnableWithExcept) Enabled(level zapcore.Level) bool {
	if !l.LevelEnabler.Enabled(level) {
		return false
	}

	return !l.except[level]
}
func newLevelEnablerWithExcept[T any](enabler zapcore.LevelEnabler, except map[zapcore.Level]T, exceptLevels ...zapcore.Level) levelEnableWithExcept {
	result := levelEnableWithExcept{
		LevelEnabler: enabler,
		except:       make(map[zapcore.Level]bool, len(except)),
	}

	for level := range except {
		result.except[level] = true
	}

	for _, exceptLevel := range exceptLevels {
		delete(result.except, exceptLevel)
	}

	return result
}
