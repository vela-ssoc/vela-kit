package vela

import (
	storm "github.com/asdine/storm/v3"
	"github.com/vela-ssoc/vela-kit/lua"
	"go.etcd.io/bbolt"
	"net"
	"os"
	"sync"
)

var (
	_G   Environment //缓存全局环境变量
	once sync.Once   //控制设置次数
)

type ScanByEnv interface {
	Scan(id, name string, chunk []byte, metadata map[string]interface{}, timeout int) error
	ScanList() []*ScanInfo
	StopScanById(id string)
	StopScanByName(name string)
	StopScanAll()
}

type CallByEnv interface {
	P(*lua.LFunction) lua.P
	Clone(*lua.LState) *lua.LState
	Coroutine() *lua.LState
	Free(*lua.LState)
	DoString(*lua.LState, string) error
	DoFile(*lua.LState, string) error
	Start(*lua.LState, lua.VelaEntry) Start //启动对象的构建
	Call(*lua.LState, *lua.LFunction, ...lua.LValue) error
}

type InjectByEnv interface {
	Set(string, lua.LValue)    //注入接口
	Global(string, lua.LValue) //全局注入接口
}

type NodeByEnv interface {
	ID() string
	Arch() string
	Inet() string
	Inet6() string
	Mac() string
	Kernel() string
	Edition() string
	LocalAddr() string
	Broker() string
	WithBroker(string, net.HardwareAddr, net.IP, net.IP, string, net.Addr) // arch , mac , inet , inet6 , edition
}

type auxiliary interface {
	Register(Closer)
	Name() string                    //当前环境的名称
	DB() *bbolt.DB                   //当前环境的缓存库
	Prefix() string                  //系统前缀
	ExecDir() string                 //当前环境目录
	Mode() string                    //当前环境模式
	IsDebug() bool                   //是否调试模式
	Spawn(int, func()) error         //异步执行 (delay int , task func())
	Notify()                         //监控退出信号
	Kill(os.Signal)                  //退出
	Bucket(...string) Bucket         //缓存
	Storm(...string) storm.Node      //storm node
	Adt() interface{}                //审计对象
	Store(string, interface{})       //存储对象
	Find(string) (interface{}, bool) //发现对象
}

type Environment interface {
	TnlByEnv
	LogByEnv
	CallByEnv
	ScanByEnv
	MimeByEnv
	NodeByEnv
	taskByEnv
	InjectByEnv
	RegionByEnv
	auxiliary
	SharedEnv
	ThirdEnv
}
