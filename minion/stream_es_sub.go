package minion

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/lua"
	"time"
)

type ElasticSubT struct {
	lua.SuperVelaData
	index  func([]byte) string
	object *stream
}

func newElasticT(L *lua.LState, object *stream) *ElasticSubT {
	index := L.CheckString(1)
	return &ElasticSubT{
		object: object,
		index: func(_ []byte) string {
			return index
		},
	}
}

func (sub *ElasticSubT) Name() string {
	return fmt.Sprintf("%s.sdk.%s", sub.object.tx.Type(), sub.index(nil))
}

func (sub *ElasticSubT) Type() string {
	return "es"
}

func (sub *ElasticSubT) Write(data []byte) (wn int, er error) {
	if !sub.object.IsRun() || sub.object.socket == nil {
		return
	}

	index := sub.index(data)
	if index == "" {
		return
	}

	if len(data) == 0 {
		return
	}

	return sub.object.socket.Write(toBulkDoc(index, data))
}

func (sub *ElasticSubT) Start() error {
	sub.V(lua.VTRun, time.Now())
	return nil
}

func (sub *ElasticSubT) Close() error {
	return nil
}

func (sub *ElasticSubT) pushL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	pn := 0
	for i := 1; i <= n; i++ {
		wn, err := sub.Write(lua.S2B(L.Get(i).String()))
		pn += wn

		if err != nil {
			L.Push(lua.LNumber(pn))
			L.Pushf("%v", err)
			return 2
		}
	}

	L.Push(lua.LInt(pn))
	L.Push(lua.LNil)
	return 2
}

func (sub *ElasticSubT) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "push":
		return lua.NewFunction(sub.pushL)

	default:
		return lua.LNil

	}
}
