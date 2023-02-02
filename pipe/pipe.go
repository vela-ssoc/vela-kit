package pipe

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/lua"
)

type Fn func(...interface{}) error

type Px struct {
	chain []Fn
	seek  int
	xEnv  vela.Environment
}

func (px *Px) clone(co *lua.LState) *lua.LState {
	if px.xEnv == nil {
		px.xEnv = vela.GxEnv()
	}

	if co == nil {
		return px.xEnv.Coroutine()
	}

	return px.xEnv.Clone(co)
}
func (px *Px) append(v Fn) {
	if v == nil {
		return
	}

	px.chain = append(px.chain, v)
}

func (px *Px) coroutine() *lua.LState {
	if px.xEnv != nil {
		return px.xEnv.Coroutine()
	}
	return vela.GxEnv().Coroutine()
}

func (px *Px) free(co *lua.LState) {
	if px.xEnv != nil {
		px.xEnv.Free(co)
		return
	}
	vela.GxEnv().Free(co)
}

func (px *Px) invalid(format string, v ...interface{}) {
	if px.xEnv == nil {
		vela.GxEnv().Errorf(format, v...)
		return
	}

	px.xEnv.Errorf(format, v...)
}

func New(opt ...func(*Px)) (px *Px) {
	px = &Px{}

	for _, fn := range opt {
		fn(px)
	}

	return
}
