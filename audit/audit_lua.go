package audit

import (
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
)

func (a *Audit) toL(L *lua.LState) int {
	a.cfg.sdk = auxlib.CheckWriter(L.CheckVelaData(1), L)
	return 0
}

func (a *Audit) pipeL(L *lua.LState) int {
	a.cfg.pipe.CheckMany(L, pipe.Seek(0))
	return 0
}

func (a *Audit) passL(L *lua.LState) int {
	key := L.CheckString(1)
	filter := L.CheckString(2)
	a.cfg.pass = append(a.cfg.pass, newFilter(key, filter))
	return 0
}

func (a *Audit) inhibitL(L *lua.LState) int {
	tag := L.CheckString(1)
	ttl := L.CheckInt(2)
	a.cfg.rate = append(a.cfg.rate, newInhibitMatch(tag, ttl))
	return 0
}

func (a *Audit) initL(L *lua.LState) int {
	adt := CheckAdt()
	cfg := newConfig(L)
	proc := L.NewVela(a.Name(), typeof)
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
	adt.init{}
	adt.pass()
	adt.pipe(_(ev) end)
	adt.to(sdk)
	adt.inhibit("$id.$inet.$from.$remote_addr.$subject" , 5 * 60) //5分钟 告警一次
*/

func (a *Audit) Index(L *lua.LState, key string) lua.LValue {
	if !checkout(L) {
		return lua.LNil
	}

	switch key {

	case "pass":
		return lua.NewFunction(a.passL)

	case "pipe":
		return lua.NewFunction(a.pipeL)

	case "to":
		return lua.NewFunction(a.toL)

	case "inhibit":
		return lua.NewFunction(a.inhibitL)

	case "start":
		return lua.NewFunction(func(co *lua.LState) int {
			xEnv.Start(L, a).From(co.CodeVM()).Do()
			return 0
		})

	default:

		//todo
		return lua.LNil
	}
}
