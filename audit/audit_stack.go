package audit

import (
	"github.com/vela-ssoc/vela-kit/execpt"
	"github.com/vela-ssoc/vela-kit/lua"
)

func RecoverByCodeVM(L *lua.LState, ev *Event) {
	r := recover()
	if r == nil {
		return
	}
	ev.Subject("进程服务异常").From(L.CodeVM()).Msg(execpt.StackTrace(0)).Log().Put()
}

func Recover(ev *Event) {
	r := recover()
	if r == nil {
		return
	}
	ev.Subject("进程异常").Msg(execpt.StackTrace(0)).Log().Put()
}
