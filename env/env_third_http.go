package env

import (
	"bytes"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"net/http"
)

func (th *third) pull(env *Environment, names []string) (thirdHttpReply, error) {
	enc := lua.Json(32)
	enc.Tab("")
	enc.Join("data", names)
	enc.End("}")

	rsp := env.HTTP(http.MethodPost, "/v1/third/infos", "", bytes.NewReader(enc.Bytes()), http.Header{
		"Content-Type": []string{"application/json"},
	})

	var reply thirdHttpReply

	if e := rsp.E(); e != nil {
		return reply, e
	}

	if e := rsp.JSON(&reply); e != nil {
		return reply, e
	}

	return reply, nil
}

func (th *third) diff(env *Environment, current map[string]*vela.ThirdInfo, reply thirdHttpReply) {
	for _, r := range reply.Data {
		info, ok := current[r.Name]
		if !ok {
			env.Errorf("not found %s third with local", r.Name)
			continue
		}
		delete(current, r.Name)

		if info.Hash != r.Hash {
			th.update(env, info, "")
			continue
		}

		if th.check(env, info) {
			th.update(env, info, "")
		}

	}
}
