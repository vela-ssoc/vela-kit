package minion

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
)

var xEnv vela.Environment

func Constructor(env vela.Environment) {
	xEnv = env
	env.Set("push", newPushEx())
	env.Set("call", lua.NewFunction(newCallL))

	uv := lua.NewUserKV()
	uv.Set("new", lua.NewFunction(newStreamSub))
	uv.Set("kfk", lua.NewFunction(newStreamKfk))
	uv.Set("tcp", lua.NewFunction(newStreamTcp))
	uv.Set("es", lua.NewFunction(newElasticL))
	env.Set("stream", uv)
}
