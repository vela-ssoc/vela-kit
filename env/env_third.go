package env

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/vela-ssoc/vela-kit/vela"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type third struct {
	dir    string
	mutex  sync.RWMutex
	cache  map[string]*vela.ThirdInfo
	bucket vela.Bucket
}

func (th *third) publish(info *vela.ThirdInfo) {
	th.mutex.Lock()
	defer th.mutex.Unlock()
	th.cache[info.Name] = info
}

func (th *third) recovery(name string) {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	delete(th.cache, name)

	tab := make(map[string]*vela.ThirdInfo, len(th.cache))
	for key, info := range th.cache {
		tab[key] = info
	}
	th.cache = tab
}

func (th *third) list() []string {
	var list []string
	th.mutex.Lock()
	defer th.mutex.Unlock()
	for key, _ := range th.cache {
		list = append(list, key)
	}
	return list
}

// thirdSyncDir 启动同步文件信息
func (th *third) thirdSyncDir(env *Environment) {
	current, _ := th.table(env)
	if len(current) == 0 {
		err := os.Remove(th.dir)
		if err != nil {
			env.Errorf("third remove %s fail %v", th.dir, err)
		}
		return
	}

	handle := func(fileInfo *vela.ThirdInfo) {
		info, ok := current[fileInfo.Name]
		if !ok {
			th.drop(env, fileInfo)
			return
		}

		if info.IsNull() {
			th.drop(env, fileInfo)
			return
		}

		if info.Size != fileInfo.Size {
			th.update(env, fileInfo, fileInfo.Hash)
			return
		}

		if info.MTime != fileInfo.MTime {
			th.update(env, fileInfo, fileInfo.Hash)
			return
		}

		if !fileInfo.Compression() && fileInfo.Hash != info.Hash {
			th.update(env, fileInfo, fileInfo.Hash)
			return
		}

		th.publish(info)
	}

	dir, err := os.ReadDir(th.dir)
	if err != nil {
		env.Errorf("third sync %s file info fail %v", th.dir, err)
		th.remove(env, current)
		return
	}

	for _, entry := range dir {
		info := &vela.ThirdInfo{
			Name: entry.Name(),
		}

		if entry.IsDir() {
			info.Name = info.Name + ".zip"
		}

		ff, e := entry.Info()
		if e != nil {
			env.Errorf("third %s read file info fail %v", info.File(), e)
			handle(info)
			continue
		}

		info.MTime = ff.ModTime().Unix()
		info.Size = ff.Size()

		hash, e := info.CheckSum()
		if e != nil {
			env.Errorf("third %s read file hash read fail %v", info.File(), e)
			handle(info)
			continue
		}

		info.Hash = hash
		info.Extension = filepath.Ext(info.Name)

		handle(info)
	}

}

func (th *third) info(name string) (*vela.ThirdInfo, bool) {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	info, ok := th.cache[name]
	return info, ok
}

func (th *third) table(env *Environment) (map[string]*vela.ThirdInfo, []string) {
	var names []string
	current := make(map[string]*vela.ThirdInfo, 32)
	th.bucket.ForEach(func(name string, data []byte) {
		names = append(names, name)
		info := &vela.ThirdInfo{}
		err := sonic.Unmarshal(data, info)
		if err != nil {
			env.Errorf("%s json unmarshal fail %v", name, err)
		}
		current[name] = info
	})
	return current, names
}

func (th *third) remove(env *Environment, expires map[string]*vela.ThirdInfo) {
	for _, info := range expires {
		th.drop(env, info)
	}
}

func (th *third) drop(env *Environment, info *vela.ThirdInfo) {
	//清除内存缓存
	th.recovery(info.Name)

	//删除本地缓存
	th.bucket.Delete(info.Name)

	//删除文件
	s, err := os.Stat(info.File())
	if err != nil {
		return
	}

	if s.IsDir() {
		err = os.RemoveAll(info.File())
	} else {
		err = os.Remove(info.File())
	}

	if err != nil {
		env.Errorf("%s remove %s fail info:%+v %v", info.Name, info.File(), info, err)
		return
	}
	env.Errorf("%s drop success info:%+v", info.Name, info)
}

func (th *third) clear(env *Environment) {
	current, _ := env.third.table(env)
	for _, info := range current {
		if e := os.Remove(info.File()); e != nil {
			env.Errorf("%s third %+v remove fail %v", info.Name, info, e)
		} else {
			env.Errorf("%s third %+v remove success", info.Name, info)
		}
		th.bucket.Delete(info.Name)
	}
	th.cache = make(map[string]*vela.ThirdInfo, 16)
}

func (th *third) check(env *Environment, info *vela.ThirdInfo) bool {
	file := info.File()
	if info.Expire != 0 && info.Expire <= int(time.Now().Unix()-info.CTime) {
		th.drop(env, info)
		env.Errorf("%s third store expire %+v:", file, info)
		return false
	}

	stat, err := os.Stat(file)
	if err != nil {
		env.Errorf("%s third got file stat fail %v", file, err)
		return true
	}

	size := stat.Size()
	if size != info.Size {
		env.Errorf("%s third update case file-size=%d cache-size=%d", info.Name, size, info.Size)
		return true
	}

	mtime := stat.ModTime().Unix()
	if mtime == info.MTime { //不更新
		return false
	}

	//文件发生变化

	//解压文件不校验hash
	if info.Compression() {
		return true
	}

	//校验本地hash
	hash, err := info.CheckSum()
	if err != nil {
		env.Errorf("%s third update case checksum got fail %v", info.Name, err)
		return true
	}

	//hash相等
	if hash == info.Hash {
		env.Errorf("%s third case file-mtime=%d cache-mtime=%d but hash not change %s", info.Name, mtime, info.MTime, hash)
		info.MTime = mtime
		th.success(env, info)
		return false
	}

	//更新hash 发起请求
	env.Errorf("%s third update case file-mtime=%d cache-mtime=%d file-hash=%s cache-hash=%s", info.Name, mtime, info.MTime, hash, info.Hash)
	return true
}

func (env *Environment) OnThirdSync(name string, drop bool) {
	th := env.third

	info, ok := th.info(name)
	if !ok {
		env.Debugf("%s third sync not found case got info %v", name, info)
		return
	}

	if drop {
		th.drop(env, info)
		return
	}

	th.update(env, info, info.Hash)
}

func (env *Environment) ThirdR(name string) (*vela.ThirdInfo, error) {
	th := env.third
	header, err := th.http(env, fmt.Sprintf("name=%s", name))
	if err != nil {
		return nil, err
	}

	info := &vela.ThirdInfo{
		Hash:      header.Hash,
		Name:      name,
		CTime:     time.Now().Unix(),
		Size:      int64(header.Size),
		Extension: header.Extension,
	}
	return info, info.Flush()
}

func (env *Environment) ThirdClear() error {
	env.third.clear(env)
	return nil
}

func (env *Environment) ThirdInfo(name string) *vela.ThirdInfo {
	env.third.mutex.RLock()
	defer env.third.mutex.RUnlock()
	return env.third.cache[name]
}

func (env *Environment) Third(name string) (*vela.ThirdInfo, error) {
	th := env.third

	info, ok := th.info(name)
	if !ok {
		env.Errorf("%s third update case not found", name)
		return th.load(env, name)
	}

	if info.Modified(env) {
		return th.load(env, info.Name)
	}

	return info, nil
}

func (env *Environment) initThird() {
	th := &third{
		dir:    "3rd",
		cache:  make(map[string]*vela.ThirdInfo, 32),
		bucket: env.Bucket("VELA_THIRD_INFO_DB"),
	}

	env.third = th

	env.OnConnect("third.sync", func() error {
		th.thirdSyncDir(env)
		th.sync(env)
		return nil
	})

}
