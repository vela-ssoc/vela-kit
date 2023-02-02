package env

import (
	"fmt"
	"go.etcd.io/bbolt"
	"net/url"
	"os"
	"path/filepath"
)

func (env *Environment) Exe() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}

	exe, err = filepath.Abs(exe)
	return exe, err
}

// ExecDir 获取当前运行路径
func (env *Environment) ExecDir() string {
	exe, err := env.Exe()
	if err != nil {
		fmt.Printf("ssoc client got work exe fail %v", err)
		return ""
	}

	return filepath.Dir(exe)
}

func (env *Environment) Mode() string {
	return env.tab.mode
}

func (env *Environment) IsDebug() bool {
	return env.tab.mode == "debug"
}

func (env *Environment) DB() *bbolt.DB {
	return env.bdb.db
}

func (env *Environment) Store(key string, v interface{}) {
	env.tupMutex.Lock()
	defer env.tupMutex.Unlock()
	env.tuple[key] = v
}

func (env *Environment) Find(key string) (interface{}, bool) {
	env.tupMutex.Lock()
	defer env.tupMutex.Unlock()
	v, ok := env.tuple[key]
	return v, ok
}

func (env *Environment) TunnelInfo() (string, string) {
	tnl, ok := env.Find("vela_tunnel_broker")
	if !ok {
		return "", ""
	}

	URL, ok := tnl.(*url.URL)
	if !ok {
		return "", ""
	}

	return URL.Hostname(), URL.Port()
}
