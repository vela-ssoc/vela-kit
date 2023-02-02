package node

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/lua"
)

var xEnv vela.Environment

func Constructor(env vela.Environment) {
	xEnv = env
	instance = newNode()
	env.Set("node", lua.NewFunction(newLuaNode))
}
