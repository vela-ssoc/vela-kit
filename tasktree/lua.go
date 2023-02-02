package tasktree

import (
	"github.com/vela-ssoc/vela-kit/vela"
)

var xEnv vela.Environment

func Constructor(env vela.Environment) {
	//初始化环境
	xEnv = env
	env.WithTaskTree(root)

	//注入方法
	codeLuaInjectApi(env)
	servLuaInjectApi(env)
}
