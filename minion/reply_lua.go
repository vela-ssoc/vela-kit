package minion

import (
	"github.com/bytedance/sonic"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
)

func (r *reply) String() string                         { return lua.B2S(r.Byte()) }
func (r *reply) Type() lua.LValueType                   { return lua.LTObject }
func (r *reply) AssertFloat64() (float64, bool)         { return 0, false }
func (r *reply) AssertString() (string, bool)           { return "", false }
func (r *reply) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (r *reply) Peek() lua.LValue                       { return r }

func (r *reply) ok() bool {
	return r.err == nil
}

func (r *reply) Byte() []byte {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("ok", r.ok())
	enc.KV("err", r.err)
	enc.KV("url", r.url.String())
	enc.KV("response", r.rsp)
	enc.End("}")
	return enc.Bytes()
}

func (r *reply) kind() string {
	if !r.ok() {
		return "{}"
	}
	if r.rsp.Count == 0 {
		return "{}"
	}

	chunk, _ := sonic.Marshal(r.rsp.Data)
	return lua.B2S(chunk)
}

func (r *reply) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "ok":
		return lua.LBool(r.ok())

	case "status":
		return lua.LInt(r.rsp.status)

	case "err":
		if r.err != nil {
			return lua.S2L(r.err.Error())
		}
		return lua.LNil

	case "path":
		return lua.S2L(r.url.Path)

	case "count":
		return lua.LInt(r.rsp.Count)

	case "query":
		return lua.S2L(r.url.RawQuery)

	case "kind":
		return lua.S2L(r.kind())
	}

	return r.element(key)
}

// MetaTable 中括号调用r["192.168.1.1"]
func (r *reply) MetaTable(L *lua.LState, key string) lua.LValue {
	return r.element(key)
}
