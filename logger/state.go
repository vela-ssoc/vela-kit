package logger

import (
	"errors"
	"fmt"
	"github.com/vela-ssoc/vela-kit/lua"
	"go.uber.org/zap"
	"os"
)

var (
	state             *zapState
	errRequiredOutput = errors.New("至少选择一种日志输出方式")
)

func init() {
	//初始化sate
	state = newZapState(defaultConfig())

	//Errorf("init logger %v succeed" , state.cfg )
}

type stop func() error

type zapState struct {
	lua.SuperVelaData

	cfg  *config
	stop stop

	sugar *zap.SugaredLogger
}

func (z *zapState) newSugar() {
	sugar, stopFn := newSugar(z.cfg)
	z.sugar = sugar
	z.stop = stopFn
}

func (z *zapState) Errorf(format string, args ...interface{}) {
	z.sugar.Errorf(format, args...)
}

func (z *zapState) clear() {
	if z.cfg.filename == "" {
		return
	}

	fd, e := os.OpenFile(z.cfg.filename, os.O_TRUNC, 0666)
	if e != nil {
		fmt.Printf("%s trunc fail %v", z.cfg.filename, e)
		return
	}
	defer fd.Close()
}

func newZapState(cfg *config) *zapState {
	obj := &zapState{
		cfg:  cfg,
		stop: func() error { return nil },
	}
	obj.newSugar()

	return obj
}
