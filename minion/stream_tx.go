package minion

import (
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/lua"
)

type Tx interface {
	CodeVM() string
	Type() string
	Index(*lua.LState, string) lua.LValue
	Clone(*lua.LState, *stream) *lua.VelaData
	Handle([]byte) *buffer.Byte
	Config(*lua.LState) *config
}
