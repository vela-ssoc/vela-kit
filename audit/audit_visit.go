package audit

import (
	"errors"
	"fmt"
	"github.com/vela-ssoc/vela-kit/lua"
	"time"
)

func (a *Audit) Name() string {
	return a.cfg.name
}

func (a *Audit) E(err error) {
	if err == nil {
		return
	}

	xEnv.Errorf("vela audit handle fail , error: %v", err)
}

func (a *Audit) Close() error {
	if !a.IsRun() {
		return errors.New(a.Name() + "can't close , err: is close")
	}

	a.V(lua.VTClose)
	if a.fd != nil {
		a.fd.Close()
	}

	a.cfg = velaMinConfig()
	return nil
}

func (a *Audit) Start() error {
	if a.IsRun() {
		return fmt.Errorf("%s is running", a.Name())
	}
	a.openFile()
	a.V(lua.VTRun, time.Now())
	return nil
}
