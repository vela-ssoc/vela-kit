package minion

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"net/url"
)

/*
local r = vela.call("ip/pass?have&kind=登录" , "192.168.1.1")

local call = vela.new_call("ip/pass?have&kind=登录")
call.cache(name , ttl)

vale.call("192.168.1.1").match("ok = true").ok(function() end)

end)
*/

const CallNotFound = -1

func (c *Call) String() string                         { return "vela.call.exdata" }
func (c *Call) Type() lua.LValueType                   { return lua.LTObject }
func (c *Call) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Call) AssertString() (string, bool)           { return "", false }
func (c *Call) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Call) Peek() lua.LValue                       { return c }

func (c *Call) bodyL(L *lua.LState, seek int) []string {
	n := L.GetTop()
	if n == 0 || n-seek < 1 {
		return nil
	}

	body := make([]string, 0, n)
	for i := seek + 1; i <= n; i++ {
		body = append(body, L.CheckAny(i).String())
	}
	return body
}

func (c *Call) retL(L *lua.LState, uri *url.URL, r *reply) int {
	q := uri.Query()
	switch {
	case q.Has("have"):
		L.Push(lua.LBool(r.fill() != CallNotFound))
		return 1

	case q.Has("count"):
		L.Push(lua.LInt(r.rsp.Count))
		return 1

	default:
		L.Push(r)
		return 1
	}
}

func (c *Call) doL(L *lua.LState, seek int) int {
	uri, err := url.Parse(c.uri)
	body := c.bodyL(L, seek)
	if err != nil {
		r := newReply(uri, body)
		r.err = err
		L.Push(r)
		return 1
	}
	return c.retL(L, uri, c.Do(uri, body))
}

func (c *Call) rL(L *lua.LState) int {
	return c.doL(L, 0)
}

func (c *Call) lruL(L *lua.LState) int {
	name := L.CheckString(1)
	size := L.IsInt(2)
	ttl := L.IsInt(3)
	c.bkt = xEnv.NewLRU(name, size)

	if ttl <= 0 {
		c.ttl = 1
	} else {
		c.ttl = ttl
	}

	return 0
}

func (c *Call) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "r":
		return lua.NewFunction(c.rL)
	case "cache":
		return lua.NewFunction(c.lruL)
	}
	return lua.LNil
}

func newCallL(L *lua.LState) int {
	lv := L.Get(1)
	switch lv.Type() {
	case lua.LTTable:
		tab := lv.(*lua.LTable)
		c := &Call{uri: tab.RawGetString("uri").String()}
		L.Push(c)
		return 1
	case lua.LTString:
		c := &Call{uri: L.CheckString(1)}
		return c.doL(L, 1)
	}

	L.Push(lua.LNil)
	return 1
}
