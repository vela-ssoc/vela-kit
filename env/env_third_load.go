package env

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/vela-ssoc/vela-kit/vela"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type thirdHeaderInfo struct {
	Code      int    `json:"-"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Path      string `json:"path"`
	Desc      string `json:"desc"`
	Size      int    `json:"size"`
	Extension string `json:"extension"`
}

func (thi thirdHeaderInfo) filepath() string {
	if thi.Path == "" {
		return ""
	}
	return filepath.Clean(filepath.Join("3rd", thi.Name))
}

type thirdHttpReply struct {
	Data []struct {
		Name string `json:"name"`
		Hash string `json:"hash"`
	} `json:"data"`
}

func (th *third) success(env *Environment, info *vela.ThirdInfo) {
	th.bucket.Push(info.Name, info.Byte(), 0)
	th.publish(info)
	env.Errorf("%s third update success info:%+v", info.Name, info)
}

// http 请求下载接口 name=aaa.lua&hash=123
func (th *third) http(env *Environment, query string) (thirdHeaderInfo, error) {
	var info thirdHeaderInfo

	r := env.GET("/v1/third/down", query)
	info.Code = r.StatusCode()

	if e := r.E(); e != nil {
		return info, e
	}

	if info.Code == http.StatusOK {
		supplement, _ := url.QueryUnescape(r.Header().Get("Content-Supplement"))
		err := sonic.Unmarshal([]byte(supplement), &info)
		if err != nil {
			return info, err
		}

		hash, err := r.SaveFile(info.filepath())
		if err != nil {
			return info, err
		}

		if hash != info.Hash {
			th.drop(env, &vela.ThirdInfo{Name: info.Name})
			info.Hash = hash
			err = fmt.Errorf("%s hash 不一致 header-md5:%s file-md5:%s", info.Hash, hash)
			return info, err
		}
		return info, nil
	}

	return info, nil
}

func (th *third) update(env *Environment, info *vela.ThirdInfo, checksum string) {

	header, err := th.http(env, fmt.Sprintf("name=%s&hash=%s&old=%s&mtime=%d&file=%s&size=%s", //hash=
		info.Name, checksum, info.Hash, info.MTime, info.File(), info.Size))
	if err != nil {
		th.drop(env, info)
		env.Errorf("update %s third fail %+v hash=%s", info.Name, info, checksum)
		return
	}

	switch header.Code {
	case http.StatusOK:
		info.Hash = header.Hash
		err = info.Flush()
	case http.StatusNotModified:
		err = info.FlushStat()

	case http.StatusNotFound:
		th.drop(env, info)

	default:
		th.drop(env, info)
		env.Errorf("%s update client http invalid code %d", info.Name, header.Code)
		return
	}

	if err != nil {
		th.drop(env, info)
		env.Errorf("update %s third fail %v", info.Name, err)
		return
	}

	th.success(env, info)
}

func (th *third) sync(env *Environment) {
	current, names := th.table(env)
	if len(names) == 0 {
		return
	}

	reply, err := th.pull(env, names)
	if err != nil {
		env.Errorf("%+v third sync pull info fail %v", names, err)
		return
	}
	th.diff(env, current, reply)
	th.remove(env, current)
}

func (th *third) load(env *Environment, name string) (*vela.ThirdInfo, error) {

	header, err := th.http(env, fmt.Sprintf("name=%s", name))
	if err != nil {
		return nil, err
	}

	info := &vela.ThirdInfo{
		Hash:      header.Hash,
		Name:      name,
		CTime:     time.Now().Unix(),
		Extension: header.Extension,
	}

	err = info.Flush()
	if err != nil {
		return info, err
	}

	th.success(env, info)
	return info, err
}
