package logger

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func encoder(format string, color bool) zapcore.Encoder {

	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}

	switch format {
	case FormatJson:
		c.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewJSONEncoder(c)
	default:
		if color {
			c.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(c)
	}
}

func newSugar(c *config) (*zap.SugaredLogger, stop) {

	//初始化stop函数
	stopFn := func() error {
		return nil
	}

	//必要函数和等级要求
	encode := encoder(c.format, c.color)
	fn := zap.LevelEnablerFunc(func(v zapcore.Level) bool {
		return v >= c.level
	})

	var core zapcore.Core

	//输出到文件
	if c.filename != "" {
		w := &lumberjack.Logger{
			Filename:   c.filename,
			MaxSize:    c.maxSize,
			MaxAge:     c.maxAge,
			MaxBackups: c.maxBackups,
			Compress:   c.compress,
		}
		core = zapcore.NewCore(encode, zapcore.AddSync(w), fn)
		stopFn = w.Close
	}

	//输出到前台
	if c.console {
		sync := zapcore.AddSync(os.Stderr)
		if core == nil {
			core = zapcore.NewCore(encode, sync, fn)
		} else {
			core = zapcore.NewTee(core, zapcore.NewCore(encode, sync, fn))
		}
	}

	return zap.New(core, zap.AddCallerSkip(c.skip), zap.WithCaller(c.caller)).Sugar(), stopFn
}
