package env

import (
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/env/sys"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/minion/tunnel"
	"github.com/vela-ssoc/vela-kit/tasktree"
	"github.com/vela-ssoc/vela-kit/vela"
	"go.uber.org/zap/zapcore"
	"sync"
)

type substance struct {
	region vela.Region
	task   *tasktree.TaskTree
}

type onConnectEv struct {
	name string
	todo func() error
}

type Environment struct {
	tab       *EnvL                  //lua vm environment
	rou       *routine               //routine pool cache
	bdb       *bboltDB               //bbolt database cache
	log       vela.Log               //External Log interface
	logLevel  zapcore.Level          //日志等级
	mbc       []vela.Closer          //Must be closed
	sub       *substance             //substance object cache
	tnl       *tunnel.Client         //数据传输通道
	adt       *audit.Audit           //审计模块
	bkr       broker                 //各种状态
	vhu       *variableHub           //变量状态
	shm       shared                 //共享内存
	mime      *MimeHub               //mime hub object
	third     *third                 //third 三方存储
	tupMutex  sync.Mutex             //并发锁
	tuple     map[string]interface{} //存储一些关键信息
	onConnect []onConnectEv
}

func withEnvironment(env *Environment) {

	//注入系统变量
	sys.WithEnv(env)

	//注入函数
	env.Set("go", lua.NewFunction(env.thread))

	//注入信号量
	env.Set("notify", lua.NewFunction(env.notifyL))

	//设置节点信息
	env.Set("ID", lua.NewFunction(env.nodeIDL))
	env.Set("inet", lua.NewFunction(env.inetL))
	env.Set("kernel", lua.S2L(env.Kernel()))
	env.Set("inet6", lua.NewFunction(env.inet6L))
	env.Set("mac", lua.NewFunction(env.macL))
	env.Set("addr", lua.NewFunction(env.addrL))
	env.Set("arch", lua.NewFunction(env.archL))
	env.Set("prefix", lua.S2L(env.ExecDir()))
	env.Set("broker", lua.NewFunction(env.brokerL))
	env.Set("load", lua.NewFunction(env.loadL))
	env.Set("exdata", lua.NewExport("vela.exdata.export", lua.WithIndex(env.exdataIndexL), lua.WithNewIndex(env.setExdataL)))
	env.Set("clear_third", lua.NewFunction(env.clearThirdL))

}

func Create(mode string, name string, protect bool) *Environment {
	env := &Environment{sub: &substance{}, tuple: make(map[string]any, 32)}
	env.newEnvL(mode, name, protect)
	env.openDb()
	env.newAudit()
	env.newRoutine()
	env.newMimeHub()
	env.newVariableHub()
	env.initThird()
	withEnvironment(env)
	return env
}
