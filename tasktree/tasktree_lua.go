package tasktree

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/worker"
	"gopkg.in/tomb.v2"
	"time"
)

func wakeupL(L *lua.LState) int {
	var tv time.Duration
	n := L.IsInt(1)
	if n > 0 {
		tv = time.Second * time.Duration(n)
	} else {
		tv = time.Second * 30
	}

	tomb := new(tomb.Tomb)
	kill := func() {
		tomb.Kill(fmt.Errorf("kill"))
	}

	poll := func() {
		tk := time.NewTicker(tv)
		defer tk.Stop()

		for {
			select {
			case <-tk.C:
				root.again()
			case <-tomb.Dying():
				return
			}
		}
	}

	worker := worker.New(L, "wakeup").Task(poll).Kill(kill)
	xEnv.Start(L, worker).From(L.CodeVM()).Do()
	return 0
}

// Index 获取服务对应的代码块
func (tt *TaskTree) Index(L *lua.LState, cname string) lua.LValue {
	tt.rLock()
	defer tt.rUnlock()

	cd, ok := CheckCodeVM(L)
	if !ok {
		goto done
	}

	//自己调用自己
	if cd.Key() == cname {
		audit.NewEvent("task").Subject("循环引用").
			Msg("%s loop call %s", L.CodeVM(), cname).Put()

		L.RaiseError("loop call %s", cname)
		return nil

	}

	cd.addLink(cname)

done:
	co, code := tt.GetCodeVM(cname)
	if co == nil {
		L.RaiseError("not found %s", cname)
		return lua.LNil
	}

	if code.IsReg() {
		wakeup(co, vela.INLINE)
	}

	return L.NewAnyData(code)
}

func servLuaInjectApi(env vela.Environment) {
	env.Set("wakeup", lua.NewFunction(wakeupL))
	env.Global("task", lua.NewAnyData(root))
}
