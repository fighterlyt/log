package log

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestConfig_Unmarshal(t *testing.T) {
	yamlCfg := `
service: test   # 服务名称
level: debug    # 日志级别，分别为debug,info,warn,error,fatal,panic
filePath: "a"   # 日志路径, 本地文件路径,如果为空，表示不输出到文件
timeZone: "b"   # 时区，默认defaultTimeZone,可以从https://www.zeitverschiebung.net/en/ 查询时区信息
timeLayout: "c" # 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
debug: true     # 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的
levelToPath:
  debug: debug.log
  info: info.log
rotate:
  maxSize: 200
`

	var (
		cfg     *Config
		wantCfg = &Config{
			Service:    "test",
			Level:      zapcore.DebugLevel,
			FilePath:   "a",
			TimeZone:   "b",
			TimeLayout: "c",
			Debug:      true,
			Rotate: &RotateConfig{
				MaxSize:    200,
				MaxBackups: 0,
				MaxAge:     0,
			},
			LevelToPath: map[zapcore.Level]string{
				zapcore.DebugLevel: `debug.log`,
				zapcore.InfoLevel:  `info.log`,
			},
		}
		err error
	)

	// data, _ := yaml.Marshal(wantCfg)
	// t.Log(string(data))
	cfg, err = NewConfigFromYamlData(strings.NewReader(yamlCfg))
	require.NoError(t, err, `从yaml数据解析配置`)
	require.EqualValues(t, wantCfg, cfg, `结果一致`)

	_, err = cfg.Build()
	require.Errorf(t, err, `构建错误`)
}

func TestConfig_Build(t *testing.T) {
	// debug = true

	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			// FilePath: `a`,
		}
	)

	originLogger, err := cfg.Build()
	require.NoError(t, err, `构建错误`)

	// With 添加字段
	originLogger = originLogger.With(zap.String(`a`, `x`), zap.String(`b`, "y"))
	// Debug输出可见
	originLogger.Debug(`a`)

	// 验证日志器
	infoLogger := originLogger.Derive(`提现`)
	// Debug 可见
	infoLogger.Debug(`提现可见`)
	// 设置为Info
	infoLogger = infoLogger.SetLevel(zapcore.InfoLevel)
	// Debug 不可见
	infoLogger.Debug(`提现不可见`)
	// Debug 可见
	originLogger.Debug(`origin Debug`)
	// Info 可见
	infoLogger.Info(`提现可见.Info`)
	// Warn 可见
	infoLogger.Warn(`提现可见.Warn`)

	infoLogger = infoLogger.With(zap.String(`info`, `info`))

	infoLogger.Info(`infoLogger.Info`)
	// 再次衍生
	debugLogger := infoLogger.Derive(`汇总`)
	// Debug不可见
	debugLogger.Debug(`不可见`)
	// 设置为Debug
	debugLogger = debugLogger.SetLevel(zapcore.DebugLevel)
	// Debug可见
	debugLogger.Debug(`汇总可见`)

	taskLogger := debugLogger.Start()
	taskLogger.Info(`开始`)

	// 三次衍生

	warnLogger := debugLogger.Derive(`三次`)
	warnLogger.Info(`三次`)
}

func TestJSON(t *testing.T) {
	var (
		jsonCfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			JSON:    true,
		}
	)

	originJSONLogger, err := jsonCfg.Build()

	require.NoError(t, err, `构建JSON输出`)

	// With 添加字段
	originJSONLogger = originJSONLogger.With(zap.String(`a`, `x`), zap.String(`b`, "y"))
	// Debug输出可见
	originJSONLogger.Debug(`a`)

	// 验证日志器
	infoJSONLogger := originJSONLogger.Derive(`提现`)
	// Debug 可见
	infoJSONLogger.Debug(`debug1`)
	// 设置为Info
	infoJSONLogger = infoJSONLogger.SetLevel(zapcore.InfoLevel)
	// Debug 不可见
	infoJSONLogger.Debug(`debug2`)
	// Debug 可见
	originJSONLogger.Debug(`origin Debug`)
	// Info 可见
	infoJSONLogger.Info(`infoJSONLogger.Info`)
	// Warn 可见
	infoJSONLogger.Warn(`infoJSONLogger.Warn`)

	infoJSONLogger = infoJSONLogger.With(zap.String(`info`, `info`))

	infoJSONLogger.Info(`infoJSONLogger.Info`)
	// 再次衍生
	debugJSONLogger := infoJSONLogger.Derive(`汇总`)
	// Debug不可见
	debugJSONLogger.Debug(`debug1`)
	// 设置为Debug
	debugJSONLogger = debugJSONLogger.SetLevel(zapcore.DebugLevel)
	// Debug可见
	debugJSONLogger.Debug(`debug2`)

	taskJSONLogger := debugJSONLogger.Start()
	taskJSONLogger.Info(`开始`)
}

func TestLogger_SetLevel(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)
	testLogger = testLogger.SetLevel(zapcore.WarnLevel)
	testLogger.Info(`info`)
	testLogger.Warn(`warn`)

	testLogger = testLogger.Derive(`a`)
	testLogger.Info(`info`)
	testLogger.Warn(`warn`)
}

func TestLogger_AddCallerSkip(t *testing.T) {
	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
		}
	)

	originLogger, err := cfg.Build()
	require.NoError(t, err, `构建错误`)

	logger := originLogger.Derive(`a`)

	logger.Info(`a`)
	logger = logger.AddCallerSkip(0)
	logger.Info(`a`)
}

func TestLogger_Start(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)

	testLogger = testLogger.With(zap.String(`key`, `b`))

	logger := testLogger.Start()

	logger.Info(`a`)

	logger = logger.With(zap.String(`key`, `a`))

	logger.Info(`a`)
}

func TestViper(t *testing.T) {
	data := `[conf]
service =  "test"   # 服务名称
level =  "debug"    # 日志级别，分别为debug,info,warn,error,fatal,panic
filePath =  "a"   # 日志路径, 本地文件路径,如果为空，表示不输出到文件
timeZone =  "b"   # 时区，默认defaultTimeZone,可以从https = //www.zeitverschiebung.net/en/ 查询时区信息
timeLayout =  "c" # 输出时间格式,默认为defaultTimeLayout,任何Go支持的格式都是合法的
debug =  true     # 是否调试，调试模式会输出完整的代码行信息,其他模式只会输出项目内部的
rotate.maxSize =  200
[levelToPath]
debug = "debug.log"
info = "info.log"
`
	viperCfg := viper.New()
	viperCfg.SetConfigType(`toml`)
	require.NoError(t, viperCfg.ReadConfig(strings.NewReader(data)), `读取`)

	require.EqualValues(t, `test`, viperCfg.GetString(`conf.service`))

	marshaledData, err := toml.Marshal(viperCfg.Get(`conf`))

	require.NoError(t, err)
	t.Log(string(marshaledData))

	var (
		cfg     = &Config{}
		wantCfg = &Config{
			Service:    "test",
			Level:      zapcore.DebugLevel,
			FilePath:   "a",
			TimeZone:   "b",
			TimeLayout: "c",
			Debug:      true,
			Rotate: &RotateConfig{
				MaxSize:    200,
				MaxBackups: 0,
				MaxAge:     0,
			},
			LevelToPath: map[zapcore.Level]string{
				zapcore.DebugLevel: `debug.log`,
				zapcore.InfoLevel:  `info.log`,
			},
		}
	)

	require.NoError(t, toml.Unmarshal(marshaledData, cfg))
	require.EqualValues(t, wantCfg, cfg)
}

func TestNewConfigFromToml(t *testing.T) {
	data := `
		Service = 'test'
        Level = 'debug'
        FilePath = 'a'
        TimeZone = 'b'
        TimeLayout = 'c'
        Debug = true
        Dev = false
        JSON = false
        HideConsole = false
        
        [Rotate]
        MaxSize = 200
        MaxBackups = 0
        MaxAge = 0
`
	var (
		cfg     = &Config{}
		wantCfg = &Config{
			Service:    "test",
			Level:      zapcore.DebugLevel,
			FilePath:   "a",
			TimeZone:   "b",
			TimeLayout: "c",
			Debug:      true,
			Rotate: &RotateConfig{
				MaxSize:    200,
				MaxBackups: 0,
				MaxAge:     0,
			},
		}
		err error
	)

	tstData, _ := toml.Marshal(wantCfg)

	t.Log(string(tstData))
	cfg, err = NewConfigFromToml([]byte(data))
	require.NoError(t, err, `读取`)
	require.EqualValues(t, wantCfg, cfg)
}

func TestLogger_AddLogrus(t *testing.T) {
	var (
		cfg = &Config{
			Service: "test",
			Level:   zapcore.DebugLevel,
			Debug:   true,
			LevelToPath: map[zapcore.Level]string{
				zapcore.DebugLevel: `debug.log`,
				zapcore.InfoLevel:  `info.log`,
			},
			Rotate: &RotateConfig{},
			// FilePath: `a`,
		}
	)

	originLogger, err := cfg.Build()

	require.NoError(t, err, `构建基础`)

	originLogger.Info(`info`)

	originLogger.Debug(`debug`)

	derived := originLogger.Derive(`derive`)

	derived.Info(`info`)
	derived.Derive(`debug`)

	// 再次derive
	derived = derived.Derive(`derive`)

	derived.Info(`info`)
	derived.Derive(`debug`)

}
