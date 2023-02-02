package audit

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
)

func CheckAdt() *Audit {
	return xEnv.Adt().(*Audit)
}

func checkout(L *lua.LState) bool {
	if !L.CheckCodeVM("audit") {
		L.RaiseError("audit not allow with %s", L.CodeVM())
		return false
	}

	return true
}

func newAdtL(L *lua.LState) int {
	if !checkout(L) {
		return 0
	}

	adt := CheckAdt()
	cfg := newConfig(L)
	proc := L.NewVela(adt.Name(), typeof)
	if proc.IsNil() {
		adt.cfg = cfg
		proc.Set(adt)
	} else {
		adt.cfg = cfg
	}

	L.Push(proc)
	return 1
}

/*
	local adt = audit.new{
		file = "xxx"
	}

	adt.to(lua.writer)

	adt.pass("id" , "*helo")

	adt.pipe(_(ev) {
	})

	adt.pipe(service.a.kfk)

	adt.start()
*/

func Constructor(env vela.Environment) {
	xEnv = env
	adt := lua.NewUserKV()
	adt.Set("ev", lua.NewFunction(newLuaEvent))
	adt.Set("event", lua.NewFunction(newLuaEvent))
	adt.Set("new", lua.NewFunction(newAdtL))
	xEnv.Set("adt", adt)

	xEnv.Set("event", lua.NewFunction(newLuaEvent))
	xEnv.Set("Debug", lua.NewFunction(newLuaDebug))
	xEnv.Set("Error", lua.NewFunction(newLuaError))
	xEnv.Set("ERR", lua.NewFunction(newLuaError))
	xEnv.Set("Info", lua.NewFunction(newLuaInfo))
	xEnv.Set("T", lua.NewFunction(newLuaObjectType))
}
