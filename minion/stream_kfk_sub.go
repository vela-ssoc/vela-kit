package minion

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
)

type ksk struct {
	lua.SuperVelaData

	topic  string
	object *stream
}

func newKsk(L *lua.LState, object *stream) *ksk {
	topic := L.CheckString(1)
	if e := auxlib.Name(topic); e != nil {
		L.RaiseError("%s invalid topic", object.Name())
		return nil
	}

	return &ksk{
		topic:  topic,
		object: object,
	}
}

func (k *ksk) Name() string {
	return fmt.Sprintf("%s.sdk.%s", k.object.tx.Type(), k.topic)
}

func (k *ksk) Type() string {
	return "kafka"
}

func (k *ksk) Write(data []byte) (wn int, er error) {
	if !k.object.IsRun() || k.object.socket == nil {
		return
	}

	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("topic", k.topic)
	enc.Raw("data", data)
	enc.End("}")
	defer func() {
		buffer.Put(enc.Buffer())
	}()

	wn, er = k.object.socket.Write(enc.Bytes())
	return
}

func (k *ksk) Start() error {
	k.V(lua.VTRun)
	return nil
}

func (k *ksk) Close() error {
	return nil
}

func (k *ksk) pushL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	pn := 0
	for i := 1; i <= n; i++ {
		wn, err := k.Write(lua.S2B(L.Get(i).String()))
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

func (k *ksk) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "push":
		return lua.NewFunction(k.pushL)

	default:
		return lua.LNil

	}
}
