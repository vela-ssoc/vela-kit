package variable

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"os"
)

func stat(L *lua.LState) int {
	filename := L.CheckString(1)

	s, e := os.Stat(filename)
	if e != nil {
		L.Push(lua.LNil)
		L.Push(lua.S2L(e.Error()))
		return 2
	}

	r := lua.NewUserKV()
	r.Set("mtime", lua.S2L(s.ModTime().Format("2006-01-02 15:04:05")))
	r.Set("size", lua.LNumber(s.Size()))

	L.Push(r)
	return 1
}

func readOnlyItem(L *lua.LState) int {
	s := L.Get(1)
	code := L.CodeVM()

	if s.Type() != lua.LTString {
		L.RaiseError("read only must be string")
		return 0
	}

	item := newVariable(s)
	item.code = code
	item.readOnly = true

	L.Push(L.NewAnyData(item))
	return 1
}

func (h *hub) NewIndex(L *lua.LState, key string, val lua.LValue) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	vm := L.CodeVM()
	item, ok := h.dict[key]
	if !ok {
		goto store
	}

	if item.readOnly && item.code != vm {
		L.RaiseError("rock.var.%s not allow with %s", key, vm)
		return
	}

store:
	if val.Type() == lua.LTAnyData {
		ros, ok := val.(*lua.AnyData).Data.(variable)
		if ok {
			h.dict[key] = ros
			return
		}
	}

	item = newVariable(val)
	item.code = vm
	h.dict[key] = item
}

func (h *hub) Index(L *lua.LState, key string) lua.LValue {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	item, ok := h.dict[key]
	if !ok {
		return lua.LNil
	}

	return item.data
}

func x(env vela.Environment) {
	ud := lua.NewAnyData(instance)
	env.Set("var", ud)
}
