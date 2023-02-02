package vkit

import (
	awk "github.com/vela-ssoc/vela-awk"
	console "github.com/vela-ssoc/vela-console"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/bucket"
	"github.com/vela-ssoc/vela-kit/hashmap"
	"github.com/vela-ssoc/vela-kit/logger"
	"github.com/vela-ssoc/vela-kit/mime"
	"github.com/vela-ssoc/vela-kit/minion"
	"github.com/vela-ssoc/vela-kit/node"
	"github.com/vela-ssoc/vela-kit/plugin"
	"github.com/vela-ssoc/vela-kit/require"
	"github.com/vela-ssoc/vela-kit/runtime"
	"github.com/vela-ssoc/vela-kit/shared"
	"github.com/vela-ssoc/vela-kit/tasktree"
	"github.com/vela-ssoc/vela-kit/thread"
	"github.com/vela-ssoc/vela-kit/vela"
)

func (dly *Deploy) withAll(xEnv vela.Environment) {
	if !dly.all {
		return
	}
	vela.WithEnv(xEnv)
	console.WithEnv(xEnv)
	awk.WithEnv(xEnv)
}

func (dly *Deploy) with(xEnv vela.Environment) {
	if dly.use == nil {
		return
	}
	dly.use(xEnv)
}

func (dly *Deploy) define() func(vela.Environment) {
	return func(xEnv vela.Environment) {
		//default inject module
		logger.Constructor(xEnv)
		runtime.Constructor(xEnv)
		mime.Constructor(xEnv)
		tasktree.Constructor(xEnv)
		plugin.Constructor(xEnv)
		bucket.Constructor(xEnv)
		node.Constructor(xEnv)
		shared.Constructor(xEnv)
		require.Constructor(xEnv)
		hashmap.Constructor(xEnv)
		thread.Constructor(xEnv)
		audit.Constructor(xEnv)
		minion.Constructor(xEnv)

		vela.WithEnv(xEnv)

		//all
		dly.withAll(xEnv)

		//custom injection
		dly.with(xEnv)
	}
}
