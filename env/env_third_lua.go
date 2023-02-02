package env

import "github.com/vela-ssoc/vela-kit/lua"

func (env *Environment) thirdSyncL(L *lua.LState) int {
	name := L.CheckString(1)
	drop := L.IsTrue(2)
	env.OnThirdSync(name, drop)
	return 0
}

func (env *Environment) clearThirdL(L *lua.LState) int {
	env.third.clear(env)
	return 0
}

func (env *Environment) loadL(L *lua.LState) int {
	name := L.CheckString(1)
	if len(name) == 0 {
		L.TypeError(1, lua.LTString)
		return 0
	}

	info, err := env.Third(name)
	if err == nil {
		L.Push(info)
		return 1
	}

	L.Push(lua.LNil)
	L.Push(lua.S2L(err.Error()))
	return 2
}
