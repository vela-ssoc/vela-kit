package audit

import (
	"encoding/json"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/opcode"
	"github.com/vela-ssoc/vela-kit/vela"
	"os"
	"reflect"
	"sync"
)

var (
	once sync.Once
	xEnv vela.Environment
)

var typeof = reflect.TypeOf((*Audit)(nil)).String()

type Audit struct {
	lua.SuperVelaData
	cfg *config
	fd  *os.File
}

func withConfig(cfg *config) *Audit {
	adt := &Audit{cfg: cfg}
	adt.V(lua.VTInit, typeof)
	return adt
}

func New() *Audit {
	return withConfig(velaMinConfig())
}

func (a *Audit) openFile() {
	fd, err := os.OpenFile(a.cfg.file, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		xEnv.Errorf("%s open file error %v", a.Name(), err)
		return
	}

	a.fd = fd
}

func (a *Audit) output(ev *Event) {
	if a.cfg.sdk != nil {
		a.cfg.sdk.Write(ev.Byte())
	}

	if a.fd != nil {
		a.fd.Write(ev.Byte())
		a.fd.Write([]byte("\n"))
	}
}

func (a *Audit) pass(ev *Event) bool {
	n := len(a.cfg.pass)
	if n == 0 {
		return false
	}

	for i := 0; i < n; i++ {
		if a.cfg.pass[i](ev) {
			return true
		}
	}

	return false
}

func (a *Audit) inhibit(ev *Event) {
	if !ev.alert {
		return
	}

	n := len(a.cfg.rate)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		if a.cfg.rate[i](a.cfg.bkt, ev) {
			ev.alert = false
			return
		}
	}
}

func (a *Audit) handle(ev *Event) {
	a.output(ev)
	if a.pass(ev) {
		xEnv.Debugf("by pass ev %s %s %s", ev.from, ev.typeof, ev.msg)
		return
	}

	//告警限速
	if ev.alert && !xEnv.IsDebug() {
		a.inhibit(ev)
	}

	//流处理
	a.cfg.pipe.Do(ev, a.cfg.co, func(err error) {
		xEnv.Errorf("%v", err)
	})

	//是否上传
	if !ev.upload {
		return
	}

	err := xEnv.TnlSend(opcode.OpEvent, json.RawMessage(ev.Byte()))
	if err != nil {
		xEnv.Errorf("%s tnl send event fail %v", xEnv.TnlName(), err)
		return
	}
	//xEnv.Debugf("%s tnl send %v event succeed", xEnv.TnlName(), ev)
}
