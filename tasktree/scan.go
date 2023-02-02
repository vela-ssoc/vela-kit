package tasktree

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"sync"
	"time"
)

type ScanPool struct {
	mu    sync.Mutex
	dirty map[string]*scanTask
}

func (sp *ScanPool) StopAll() {
	for _, task := range sp.dirty {
		task.StopScanTask()
	}
}

func (sp *ScanPool) List() []*vela.ScanInfo {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	n := len(sp.dirty)
	tuple := make([]*vela.ScanInfo, 0, n)
	for key, task := range sp.dirty {
		info := &vela.ScanInfo{
			Name:    key,
			Link:    task.code.Link(),
			Hash:    task.code.Hash(),
			From:    task.code.From(),
			Status:  task.code.Status(),
			Uptime:  task.code.header.uptime,
			ID:      task.code.header.id,
			Dialect: task.code.header.dialect,
		}
		info.Runners = task.code.List()

		if task.code.header.err == nil {
			info.Failed = false
		} else {
			info.Cause = task.code.header.err.Error()
		}

		tuple = append(tuple, info)
	}
	return tuple
}

func (sp *ScanPool) Get(name string) *scanTask {
	s, ok := sp.dirty[name]
	if !ok {
		return nil
	}

	return s
}

func newScanPool() *ScanPool {
	return &ScanPool{
		dirty: make(map[string]*scanTask, 32),
	}
}

func (tt *TaskTree) StopScanAll() {
	tt.spool.StopAll()
}

func (tt *TaskTree) StopScanById(id string) {
	tt.spool.mu.Lock()
	defer tt.spool.mu.Unlock()

	scan := tt.spool.dirty[id]
	if scan == nil {
		return
	}
	scan.StopScanTask()
}

func (tt *TaskTree) StopScanByName(name string) {
	tt.spool.mu.Lock()
	defer tt.spool.mu.Unlock()

	for _, scan := range tt.spool.dirty {
		if scan.code.Key() != name {
			continue
		}

		scan.StopScanTask()
	}
}

func (tt *TaskTree) ScanList() []*vela.ScanInfo {
	return tt.spool.List()
}

func (tt *TaskTree) Scan(env vela.Environment, id string, cname string, chunk []byte,
	metadata map[string]interface{}, timeout int) error {
	tt.spool.mu.Lock()
	defer tt.spool.mu.Unlock()

	if s := tt.spool.Get(cname); s != nil {
		s.StopScanTask()
		time.Sleep(3 * time.Second)
	}

	task := newScanTask(env, id, cname, chunk, metadata, timeout)
	tt.spool.dirty[id] = task
	return task.call(env)
}
