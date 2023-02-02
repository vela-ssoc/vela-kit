package logger

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
)

var xEnv vela.Environment

func reload(obj *zapState) {
	old := *state
	state = obj

	old.sugar.Sync()
	old.stop()

	xEnv.WithLogger(obj.sugar)
	xEnv.WithLevel(obj.cfg.level)
}

func constructor(L *lua.LState) int {
	if L.CodeVM() != "startup" {
		L.RaiseError("new logger not allowed in %s", L.CodeVM())
		return 0
	}

	cfg := newConfig(L)
	//重启
	reload(newZapState(cfg))
	return 0
}

func newClearLoggerL(L *lua.LState) int {
	state.clear()
	return 0
}

func Constructor(env vela.Environment) {
	//初始化
	xEnv = env
	xEnv.WithLogger(state.sugar)

	//日志配置
	env.Set("logger", lua.NewFunction(constructor))
	env.Set("clear_logger", lua.NewFunction(newClearLoggerL))
}
