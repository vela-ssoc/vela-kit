package minion

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"net/http"
	"net/url"
	"strings"
)

type quality []string

func (q quality) have(v string) int {
	if q.len() <= 0 {
		return -1
	}

	for i, item := range q {
		if item == v {
			return i
		}
	}

	return -1
}

func (q quality) len() int {
	return len(q)
}

func (q quality) get(sp int) string {
	if sp < 0 || sp > q.len() {
		return ""
	}
	return q[sp]
}

func (q quality) String() string                         { return strings.Join(q, ",") }
func (q quality) Type() lua.LValueType                   { return lua.LTObject }
func (q quality) AssertFloat64() (float64, bool)         { return 0, false }
func (q quality) AssertString() (string, bool)           { return "", false }
func (q quality) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (q quality) Peek() lua.LValue                       { return q }

func (q *quality) Index(L *lua.LState, key string) lua.LValue {
	val := q.have(key)
	return lua.LNumber(val)
}

func (q *quality) MetaTable(L *lua.LState, key lua.LValue) lua.LValue {
	switch key.Type() {
	case lua.LTNil:
		return lua.LNil
	case lua.LTInt:
		val := q.get(int(key.(lua.LInt)) - 1)
		if val == "" {
			return lua.LNil
		}
		return lua.S2L(val)

	case lua.LTNumber:
		val := q.get(int(key.(lua.LNumber)) - 1)
		if val == "" {
			return lua.LNil
		}
		return lua.S2L(val)

	case lua.LTString:
		return q.Index(L, key.String())
	}

	return lua.LNil
}

type response struct {
	status int                `json:"-"`
	header http.Header        `json:"-"`
	Count  int                `json:"count"`
	Data   map[string]quality `json:"data"`
}

type reply struct {
	err  error
	body []string
	url  *url.URL
	rsp  response
}

func (r *reply) element(key string) quality {
	if !r.ok() {
		return nil
	}

	return r.rsp.Data[key]
}

func (r *reply) have(key, tag string) int {
	if !r.ok() {
		return -1
	}

	q := r.element(key)
	if q == nil {
		return -1
	}

	return q.have(tag)
}

func (r *reply) fill() int {
	if !r.ok() {
		return CallNotFound
	}

	for i, el := range r.body {
		if r.element(el) != nil {
			return i
		}
	}

	return CallNotFound
}

func newReply(uri *url.URL, body []string) *reply {
	return &reply{url: uri, body: body}
}
