package env

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"sync"
)

type variableHub struct {
	mutex sync.RWMutex
	dict  map[string]variable
}

func newVariableHub() *variableHub {
	return &variableHub{
		dict: make(map[string]variable, 16),
	}
}

func (h *variableHub) find(key string, v *variable) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	item, ok := h.dict[key]
	if ok {
		*v = item
		return ok
	}
	return false
}

func (env *Environment) newVariableHub() {
	env.vhu = newVariableHub()
	env.Set("var", lua.NewAnyData(env.vhu))
	env.Set("readonly", lua.NewFunction(readOnlyItem))
}
