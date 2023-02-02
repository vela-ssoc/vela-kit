package tasktree

import (
	"errors"
	"fmt"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/execpt"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/pcall"
	"github.com/vela-ssoc/vela-kit/vela"
	tomb "gopkg.in/tomb.v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vela-ssoc/vela-kit/lua"
)

var (
	NotFoundCode = errors.New("not found code")
	root         = &TaskTree{spool: newScanPool()}
)

type TaskTree struct {
	//控制速率
	ss uint32

	//服务信息
	tom tomb.Tomb

	//锁
	mu   sync.RWMutex
	pool []*lua.LState

	spool *ScanPool
	//wakeup state
	atomicWakeup uint32
}

func (tt *TaskTree) rLock() {
	tt.mu.RLock()
}

func (tt *TaskTree) rUnlock() {
	tt.mu.RUnlock()
}

func (tt *TaskTree) lock() {
	tt.mu.Lock()
}

func (tt *TaskTree) unLock() {
	tt.mu.Unlock()
}

func (tt *TaskTree) remove(key string) error {
	idx := -1
	var link []string

	for i := 0; i < tt.Len(); i++ {
		code := tt.CodeVM(i)
		if code.inLink(key) {
			link = append(link, code.Key())
			continue
		}

		if code.Key() == key {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.New("not found " + key + " code")
	}

	if len(link) != 0 {
		return errors.New("find " + key + " code link in " + strings.Join(link, ","))
	}

	tt.del(idx)
	return nil
}

func (tt *TaskTree) CodeVM(idx int) *Code {
	code, _ := CheckCodeVM(tt.pool[idx])
	return code
}

func (tt *TaskTree) Close() error {
	tt.tom.Kill(fmt.Errorf("close"))

	tt.lock()
	defer tt.unLock()

	for i := 0; i < tt.Len(); i++ {
		code := tt.CodeVM(i)
		pcall.Exec(code.Close).Time(time.Millisecond).Spawn()
	}

	return nil
}

func (tt *TaskTree) Byte() []byte {
	tt.rLock()
	defer tt.rUnlock()

	buf := kind.NewJsonEncoder()
	buf.Arr("")

	n := len(tt.pool)
	for i := 0; i < n; i++ {
		cd := tt.CodeVM(i)
		buf.Tab("")
		buf.KV("key", cd.Key())
		buf.KV("link", cd.Link())
		buf.KV("status", cd.Status())

		buf.KV("hash", cd.Hash())
		buf.KV("way", cd.From())
		buf.KV("uptime", cd.Time())

		if e := cd.Wrap(); e != nil {
			buf.KV("err", e.Error())
		} else {
			buf.KV("err", nil)
		}

		buf.Arr("vela")
		cd.foreach(func(name string, ud *lua.VelaData) bool {
			buf.Tab("")
			buf.KV("name", name)
			buf.KV("type", ud.Data.Type())
			buf.KV("state", ud.Data.State().String())
			buf.End("},")
			return true
		})
		buf.End("]},")
	}

	buf.End("]")
	return buf.Bytes()
}

func (tt *TaskTree) String() string {
	return lua.B2S(tt.Byte())
}

func (tt *TaskTree) ForEach(fn func(idx int, code *Code) bool) {
	tt.rLock()
	defer tt.rUnlock()

	for i := 0; i < tt.Len(); i++ {
		code := tt.CodeVM(i)
		if !fn(i, code) {
			return
		}
	}
}

func (tt *TaskTree) GetCodeVM(name string) (*lua.LState, *Code) {
	tt.rLock()
	defer tt.rUnlock()

	for i := 0; i < tt.Len(); i++ {
		co := tt.pool[i]
		code := tt.CodeVM(i)
		if code.Key() == name {
			return co, code
		}
	}
	return nil, nil
}

func (tt *TaskTree) code(name string) *Code {
	_, code := tt.GetCodeVM(name)
	return code
}

func (tt *TaskTree) insert(co *lua.LState) {
	//加锁
	tt.mu.Lock() //不能取消

	//设置对象池
	tt.pool = append(tt.pool, co)

	//解锁
	tt.mu.Unlock()
}

func (tt *TaskTree) Keys(line string) []string {
	tt.rLock()
	defer tt.rUnlock()

	var keys []string

	for i := 0; i < tt.Len(); i++ {
		code := tt.CodeVM(i)
		name := code.Key()

		if line == "" {
			keys = append(keys, name)
			continue
		}

		if strings.HasPrefix(name, line) {
			keys = append(keys, name)
		}
	}
	return keys
}

func (tt *TaskTree) del(idx int) {
	tt.lock()
	defer tt.unLock()

	cd := tt.pool[idx]
	if idx == tt.Len()-1 {
		tt.pool = tt.pool[:idx]
		goto done
	}
	tt.pool = append(tt.pool[:idx], tt.pool[idx+1:]...)

done:
	defer cd.Close()
	code, _ := CheckCodeVM(cd)
	freeCodeVM(code)
}

func (tt *TaskTree) reg(id, cname string, chunk []byte, env vela.Environment, way vela.Way, dialect bool) *lua.LState {
	co, code := tt.GetCodeVM(cname)
	if co == nil {
		co, code = newCodeVM(cname, chunk, env, way)
		tt.insert(co)
		newCodeEv(code, "添加服务").Msg("%tt 注册成功", cname)
		goto done
	}

	code.ToUpdate()
	code.Update(co, chunk, env, way)
	newCodeEv(code, "更新服务").Msg("%tt 注册成功", cname)

done:
	code.header.id = id
	code.header.dialect = dialect

	return co
}

func (tt *TaskTree) wakeup(way vela.Way) error {
	tt.rLock()
	defer tt.rUnlock()

	errs := execpt.New()
	for _, co := range tt.pool {
		code, _ := CheckCodeVM(co)
		errs.Try(code.Key(), wakeup(co, way))
	}
	return errs.Wrap()
}

// load console代码
func (tt *TaskTree) load(key string, chunk []byte, env vela.Environment, sess interface{}) error {
	co := tt.reg("", key, chunk, env, vela.CONSOLE, true)
	defer func() {
		co.SetMetadata(0, nil)
	}()
	co.SetMetadata(0, sess)
	return wakeup(co, vela.CONSOLE)
}

// ExistProc 判断服务对象是否存在
func (tt *TaskTree) ExistProc(proc string) bool {
	ret := false
	tt.ForEach(func(idx int, code *Code) bool {
		if code.Exist(proc) {
			ret = true
			return false //终止 foreach
		}
		return true
	})

	return ret
}

// ExistCode 判断服务代码快是否存在
func (tt *TaskTree) ExistCode(key string) bool {
	ret := false
	tt.ForEach(func(idx int, code *Code) bool {
		if code.Key() == key {
			ret = true
			return false //终止 forearch
		}
		return true
	})
	return ret
}

func (tt *TaskTree) again() {
	w := atomic.AddUint32(&root.atomicWakeup, 1)
	if w != 1 {
		audit.NewEvent("task.again").Subject("定时任务").Msg("同步服务状态").Put()
		return
	}

	n := root.Len()
	for i := 0; i < n; i++ {
		wakeup(root.pool[i], vela.AGAIN)
	}
	atomic.StoreUint32(&root.atomicWakeup, 0)
}
