package vkit

import (
	"github.com/vela-ssoc/vela-kit/agent"
	"github.com/vela-ssoc/vela-kit/env"
	"github.com/vela-ssoc/vela-kit/vela"
)

type Deploy struct {
	name string
	all  bool
	use  func(env vela.Environment)
}

type EngineFunc func(*Deploy)

func All() EngineFunc {
	return func(e *Deploy) {
		e.all = true
	}
}

func Use(fn func(vela.Environment)) EngineFunc {
	return func(e *Deploy) {
		e.use = fn
	}
}

func New(name string, options ...EngineFunc) *Deploy {
	e := &Deploy{name: name}
	for _, fn := range options {
		fn(e)
	}
	return e
}

func (dly *Deploy) Agent() {
	agent.By(dly.name, dly.define())
}

func (dly *Deploy) Debug(hide Hide) {
	xEnv := env.Create("debug", dly.name, hide.Protect)
	dly.define()(xEnv)

	xEnv.Error("ssc sensor debug start")
	xEnv.Spawn(0, func() {
		xEnv.Dev(hide.Lan, hide.Vip, hide.Edition, hide.Hostname)
	})

	xEnv.Error("ssc sensor debug succeed")
	xEnv.Notify()
	xEnv.Error("ssc sensor exit succeed")
}
