package env

import (
	"github.com/vela-ssoc/vela-kit/lua"
)

type variable struct {
	code     string
	data     lua.LValue
	readOnly bool
}

func (v variable) String() string                         { return v.code }
func (v variable) Type() lua.LValueType                   { return lua.LTObject }
func (v variable) AssertFloat64() (float64, bool)         { return 0, false }
func (v variable) AssertString() (string, bool)           { return "", false }
func (v variable) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (v variable) Peek() lua.LValue                       { return v }

func newVariable(val lua.LValue) variable {
	if val.Type() != lua.LTObject {
		return variable{data: val, readOnly: false}

	}

	v, is := val.(variable)
	if is {
		return v
	}

	return variable{data: val, readOnly: false}
}

func (v variable) ReadOnly(L *lua.LState, key string) {
	vm := L.CodeVM()
	if v.readOnly && v.code != vm {
		L.RaiseError("rock.var.%s not allow with %s", key, vm)
	}
}

func readOnlyItem(L *lua.LState) int {
	val := L.CheckString(1)
	code := L.CodeVM()

	item := newVariable(lua.S2L(val))
	item.code = code
	item.readOnly = true
	L.Push(item)
	return 1
}

func (h *variableHub) NewIndex(L *lua.LState, key string, val lua.LValue) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	vm := L.CodeVM()
	item, ok := h.dict[key]
	if ok {
		item.ReadOnly(L, key)
	}

	item = newVariable(val)
	item.code = vm
	h.dict[key] = item
}

func (h *variableHub) Index(L *lua.LState, key string) lua.LValue {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	item, ok := h.dict[key]
	if !ok {
		return lua.LNil
	}

	return item.data
}
