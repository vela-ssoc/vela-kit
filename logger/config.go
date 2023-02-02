package logger

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"go.uber.org/zap/zapcore"
)

const (
	FormatJson = "json"
	FormatText = "text"

	LevelDebug  = "DEBUG"
	LevelInfo   = "INFO"
	LevelWarn   = "WARN"
	LevelError  = "PTErr"
	LevelDpanic = "DPANIC"
	LevelPanic  = "PTPanic"
	LevelFatal  = "FATAL"
)

type config struct {
	level      zapcore.Level `ini:"level" yaml:"level" json:"level"`                      // 日志输出级别
	filename   string        `ini:"filename" yaml:"filename" json:"filename"`             // 文件输出位置, 留空则代表不输出到文件
	maxSize    int           `ini:"maxSize" yaml:"maxSize" json:"maxSize"`                // 单个文件大小, 单位: MiB
	maxBackups int           `ini:"maxBackups" yaml:"maxBackups" json:"maxBackups"`       // 最大文件备份个数
	maxAge     int           `ini:"maxAge" yaml:"maxAge" json:"maxAge"`                   // 日志文件最长留存天数
	compress   bool          `ini:"compress" yaml:"compress" json:"compress"`             // 备份日志文件是否压缩
	console    bool          `ini:"vela-console" yaml:"vela-console" json:"vela-console"` // 是否输出到控制台
	caller     bool          `ini:"caller" yaml:"caller" json:"caller"`                   // 是否打印调用者
	format     string        `ini:"format" yaml:"format" json:"format"`                   // 日志格式化方式
	color      bool          `ini:"color" yaml:"color" json:"color"`                      // 是否显示颜色
	skip       int           `ini:"skip" yaml:"skip" json:"skip"`                         // 打印代码层级
}

func defaultConfig() *config {
	return &config{
		level:      zapcore.DebugLevel,
		filename:   "",
		maxSize:    1024,
		maxBackups: 1024,
		maxAge:     180,
		compress:   false,
		console:    true,
		caller:     true,
		skip:       1,
		format:     FormatText,
	}

}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := defaultConfig()

	tab.Range(func(key string, value lua.LValue) {
		cfg.NewIndex(L, key, value)
	})

	if err := cfg.verify(); err != nil {
		L.RaiseError("logger verify err: %v", err)
		return nil
	}

	return cfg
}

func (c *config) NewIndex(L *lua.LState, key string, val lua.LValue) {

	switch key {

	case "level":
		c.level = checkLevel(L, val.String())

	case "filename":
		c.filename = val.String()

	case "max_size":
		c.maxSize = lua.CheckInt(L, val)

	case "max_backups":
		c.maxBackups = lua.CheckInt(L, val)

	case "max_age":
		c.maxAge = lua.CheckInt(L, val)

	case "compress":
		c.compress = lua.CheckBool(L, val)

	case "vela-console":
		c.console = lua.CheckBool(L, val)

	case "caller":
		c.caller = lua.CheckBool(L, val)

	case "skip":
		c.skip = lua.CheckInt(L, val)

	case "format":
		c.format = val.String()
	}
}

func (c *config) verify() error {
	if c.filename == "" && !c.console {
		return errRequiredOutput
	}
	return nil
}

func checkLevel(L *lua.LState, val string) zapcore.Level {
	var level zapcore.Level

	err := level.UnmarshalText(lua.S2B(val))
	if err != nil {
		L.RaiseError("logger.level got error: %v", err)
		return 0
	}
	return level
}
