package env

import (
	"errors"
	"github.com/vela-ssoc/vela-kit/binary"
	"github.com/vela-ssoc/vela-kit/minion/dispatch"
	"github.com/vela-ssoc/vela-kit/minion/tunnel"
	opcode "github.com/vela-ssoc/vela-kit/opcode"
	"github.com/vela-ssoc/vela-kit/safecall"
	"github.com/vela-ssoc/vela-kit/vela"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	notFoundTnlE = errors.New("not found tunnel client")
)

//func (env *Environment) WithTnl(v interface{}) {
//	if v == nil {
//		env.Error("inject tunnel client fail , got nil")
//		return
//	}
//
//	tnl, ok := v.(*tunnel.Client)
//	if !ok {
//		env.Error("inject tunnel client fail , got %T", v)
//		return
//	}
//
//	env.tnl = tnl
//}

func (env *Environment) TnlName() string {
	if env.tnl == nil {
		return "nil"
	}

	return env.tnl.Name()
}

func (env *Environment) TnlVersion() string {
	if env.tnl == nil {
		return "nil"
	}

	return env.tnl.Version()
}

func (env *Environment) TnlIsDown() bool {
	if env.tnl == nil {
		return false
	}

	return env.tnl.Inactive()
}

func (env *Environment) TnlSend(op opcode.Opcode, v interface{}) error {
	if env.tnl == nil {
		return notFoundTnlE
	}

	return env.tnl.Push(op, v)
}

func (env *Environment) GET(path, query string) vela.HTTPResponse {
	if env.tnl == nil {
		return tunnel.HTTPResponse{Error: notFoundTnlE}
	}

	header := http.Header{"Content-Type": []string{"application/json"}}
	return env.HTTP(http.MethodGet, path, query, nil, header)
}

func (env *Environment) DoHTTP(r *http.Request) vela.HTTPResponse {
	if env.tnl == nil {
		return tunnel.HTTPResponse{Error: notFoundTnlE}
	}
	return env.tnl.Do(r)
}

func (env *Environment) HTTP(method string, path string, query string, body io.Reader, header http.Header) vela.HTTPResponse {
	if env.tnl == nil {
		return tunnel.HTTPResponse{Error: notFoundTnlE}
	}

	return env.tnl.HTTP(method, path, query, body, header)
}

func (env *Environment) PostJSON(path string, data interface{}, reply interface{}) error {
	if env.tnl == nil {
		return notFoundTnlE
	}

	return env.tnl.PostJSON(path, data, reply)
}

func (env *Environment) Stream(mode string, data interface{}) (vela.HTTPStream, error) {
	if env.tnl == nil {
		return nil, notFoundTnlE
	}

	return env.tnl.Stream(mode, data)
}

func (env *Environment) Dev(lan []string, vip []string, edit string, host string) {

	hide := tunnel.Hide{
		LAN:        lan,
		VIP:        vip,
		Edition:    edit,
		Servername: host,
	}

	env.tnl = tunnel.New(hide, tunnel.WithEnv(env), tunnel.WithHandler(dispatch.WithEnv(env)))
	if err := env.tnl.Start(); err != nil {
		env.Errorf("vela minion client error: %v", err)
		return
	}

	env.Errorf("vela minion client start succeed")
	env.onConnectHandler()
}

func (env *Environment) onConnectHandler() {

	for _, ev := range env.onConnect {
		env.Errorf("%s onconnect todo start", ev.name)
		go func(name string, todo func() error) {
			safecall.New(true).
				Timeout(60 * time.Second).
				OnError(func(err error) { env.Errorf("%s on connect todo exec fail %v", name, err) }).
				OnTimeout(func() { env.Errorf("%s on connect todo exec timeout", name) }).
				OnPanic(func(v interface{}) { env.Errorf("%s on connect todo exec panic %v", name, v) }).
				Exec(todo)
			env.Errorf("%s onconnect todo exec over", ev.name)
		}(ev.name, ev.todo)
	}
}

func (env *Environment) Worker() {

	// 从自身文件中取出携带的配置, 测试环境, 不作错误处理
	exe, err := env.Exe()
	if err != nil {
		env.Errorf("not found exe")
		env.Kill(os.Kill)
		return
	}

	bin, err := binary.Decode(exe)
	if err != nil {
		env.Errorf("ssc config unmarshal error %v", err)
		env.Kill(os.Kill)
		os.Exit(0)
	}

	hide := tunnel.Hide{
		LAN:        bin.LAN,
		VIP:        bin.VIP,
		Edition:    bin.Edition,
		Servername: bin.Servername,
	}

	env.tnl = tunnel.New(hide, tunnel.WithEnv(env), tunnel.WithHandler(dispatch.WithEnv(env)))
	if e := env.tnl.Start(); e != nil {
		env.Errorf("ssc cli worker over error %v", e)
		os.Exit(0)
	}

	env.onConnectHandler()
}

func (env *Environment) OnConnect(name string, fn func() error) {

	for _, ev := range env.onConnect {
		if ev.name == name {
			env.Errorf("%s On tunnel connect function already ok", name)
			return
		}
	}

	env.onConnect = append(env.onConnect, onConnectEv{
		name: name,
		todo: fn,
	})
}
