package minion

import (
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/lua"
)

/*
	local es = vela.stream.es{
		name = "ee",
		addr = "http://127.0.0.1:9200",
		index = "hello",
		username = "aaa",
		password = "aaa",
	}

*/

type Elastic struct {
	tab   *lua.LTable
	index string
	code  string
}

func newElastic(L *lua.LState) *Elastic {
	tab := L.CheckTable(1)
	es := &Elastic{tab: tab, code: L.CodeVM()}
	es.index = tab.RawGetString("index").String()
	return es
}

func (es *Elastic) Type() string {
	return "es"
}

func (es *Elastic) EsIndex(raw []byte) string {
	return es.index
}

func (es *Elastic) Handle(raw []byte) *buffer.Byte {
	if len(raw) == 0 {
		return nil
	}

	index := es.EsIndex(raw)
	if index == "" {
		return nil
	}

	return &buffer.Byte{B: toBulkDoc(es.index, raw)}
}

func (es *Elastic) Config(L *lua.LState) *config {
	return newConfig(L, es.tab)
}

func (es *Elastic) CodeVM() string {
	return es.code
}

func (es *Elastic) Clone(L *lua.LState, s *stream) *lua.VelaData {
	return lua.NewVelaData(newElasticT(L, s))
}

func (es *Elastic) indexL(L *lua.LState) int {
	return 0
}

func (es *Elastic) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "index":
		return lua.NewFunction(es.indexL)
	}
	return lua.LNil
}

func newElasticL(L *lua.LState) int {
	return help(L, newElastic(L))
}
