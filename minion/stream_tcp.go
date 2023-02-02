package minion

import (
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/lua"
)

type tcp struct {
	tab  *lua.LTable
	code string
}

func newTcp(L *lua.LState) *tcp {
	tab := L.CheckTable(1)
	tab.RawSetString("network", lua.S2L("tcp"))
	return &tcp{tab: tab, code: L.CodeVM()}
}

func (t *tcp) Type() string {
	return "tcp.forward"
}
func (t *tcp) Handle(raw []byte) *buffer.Byte {
	return &buffer.Byte{B: raw}
}

func (t *tcp) Config(L *lua.LState) *config {
	return newConfig(L, t.tab)
}

func (t *tcp) CodeVM() string {
	return t.code
}

func (t *tcp) Clone(L *lua.LState, s *stream) *lua.VelaData {
	return nil
}

func (t *tcp) Index(L *lua.LState, key string) lua.LValue {
	return lua.LNil
}

func newStreamTcp(L *lua.LState) int {
	return help(L, newTcp(L))
}
