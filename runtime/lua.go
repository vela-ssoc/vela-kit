package runtime

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/lua"
	"runtime"
)

var xEnv vela.Environment

func Constructor(env vela.Environment) {
	xEnv = env

	rtv := lua.NewUserKV()
	rtv.Set("code", lua.NewFunction(codeL))
	rtv.Set("free", lua.NewFunction(freeL))

	rtv.Set("max_memory", lua.NewFunction(setMaxMemoryL))
	rtv.Set("max_thread", lua.NewFunction(setMaxThreadL))
	rtv.Set("max_cpu", lua.NewFunction(setMaxCpuL))
	rtv.Set("memory", lua.NewFunction(memoryL))
	rtv.Set("p_memory", lua.NewFunction(pMemoryL))

	rtv.Set("pprof", lua.NewFunction(pprofL))

	rtv.Set("OS", lua.S2L(runtime.GOOS))
	rtv.Set("ARCH", lua.S2L(runtime.GOARCH))
	rtv.Set("CPU_CORE", lua.LInt(runtime.NumCPU()))
	rtv.Set("windows", lua.LBool(goos == "windows"))
	rtv.Set("linux", lua.LBool(goos == "linux"))
	env.Global("runtime", rtv)

}
