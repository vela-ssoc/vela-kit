package tasktree

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/xreflect"
	"github.com/vela-ssoc/vela-kit/lua"
)

func startL(L *lua.LState) int {
	n := L.GetTop()
	for i := 1; i <= n; i++ {
		ud := L.CheckVelaData(i)
		xEnv.Start(L, ud.Data).From(ud.CodeVM()).Do()
	}
	return 0
}

func privateL(L *lua.LState) int {

	cname := L.CodeVM()
	if cname == "" {
		L.RaiseError("not allow inline , must be code vm")
		return 0
	}

	n := L.GetTop()
	for i := 1; i <= n; i++ {
		ud := L.CheckVelaData(i)
		ud.Private(L)
		for _, co := range root.pool {
			code, _ := CheckCodeVM(co)
			if !code.inLink(cname) {
				continue
			}

			code.ToUpdate()
			xEnv.Errorf("%s code with inline , %s set update reg", cname, code.Key())
		}
	}
	return 0
}

// Index CODE结果中的PROC服务
func (cd *Code) Index(L *lua.LState, key string) lua.LValue {
	ud := cd.vela(key)
	if ud == nil {
		L.RaiseError("not found %s vela", key)
		return lua.LNil
	}

	if !ud.IsPrivate() {
		return ud
	}

	if !cd.CompareVM(L) {
		L.RaiseError("%s link %s inline vela", L.CodeVM(), ud.Data.Name())
		return lua.LNil
	}

	return ud

}

func metadataL(L *lua.LState) int {
	co, ok := CheckCodeVM(L)
	if !ok {
		return 0
	}

	key := L.CheckString(1)

	L.Push(xreflect.ToLValue(co.metadata[key], L))
	return 1
}

func codeLuaInjectApi(env vela.Environment) {
	env.Global("start", lua.NewFunction(startL))
	env.Global("private", lua.NewFunction(privateL))
	env.Global("metadata", lua.NewFunction(metadataL))
}
