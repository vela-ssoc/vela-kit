package variable

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"sync"
)

var (
	instance *hub
)

func init() {
	instance = &hub{
		mutex: sync.RWMutex{},
		dict:  make(map[string]variable, 32),
	}
}

type variable struct {
	code     string
	data     lua.LValue
	readOnly bool
}

type hub struct {
	mutex sync.RWMutex
	dict  map[string]variable
}

func newVariable(v lua.LValue) variable {
	return variable{data: v, readOnly: false}
}

func (h *hub) find(key string, v *variable) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	item, ok := h.dict[key]
	if ok {
		*v = item
		return ok
	}
	return false
}
