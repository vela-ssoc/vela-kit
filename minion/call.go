package minion

import (
	"bytes"
	"fmt"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/vela"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

/*

	{
		"uri": "ip/attack",
		"data":["192.168.1.1" , "192.168.1.2"]
	}

	{
		"uri": "",
		"count": 0,
		"message": map[string][]string
	}
*/

var (
	JsonHeader = http.Header{
		"Content-Type": []string{"application/json"},
	}
)

type Call struct {
	ttl int
	uri string
	bkt vela.Shared
}

func (c *Call) bodyToReader(body []string) io.Reader {
	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.Join("data", body)
	enc.End("}")
	return bytes.NewReader(enc.Bytes())
}

func (c *Call) POST(path, query string, body []string, r *reply) {
	rsp := xEnv.HTTP(http.MethodPost, path, query, c.bodyToReader(body), JsonHeader)
	if e := rsp.E(); e != nil {
		r.err = e
		return
	}

	if e := rsp.JSON(&r.rsp); e != nil {
		r.err = e
		return
	}

	r.rsp.status = rsp.StatusCode()
	r.rsp.header = rsp.Header()
}

func (c *Call) Cache(uri *url.URL) vela.Shared {
	if c.bkt != nil {
		return c.bkt
	}

	name := uri.Query().Get("cache")
	ttl := uri.Query().Get("ttl")
	n, _ := strconv.Atoi(ttl)
	if ttl == "" || n == 0 {
		c.bkt = xEnv.NewLRU(name, 30)
	} else {
		c.bkt = xEnv.NewLRU(ttl, n)
	}

	return c.bkt
}

func (c *Call) One(uri *url.URL, path string, body []string, r *reply) {
	line := body[0]
	if c.bkt == nil {
		c.POST(path, uri.RawQuery, body, r)
		return
	}

	v, _ := c.bkt.Get(line)
	if info, ok := v.(*reply); ok {
		r.err = info.err
		r.rsp = info.rsp
		return
	}

	c.POST(path, uri.RawQuery, body, r)
	err := c.bkt.Set(line, r, c.ttl)
	if err != nil {
		xEnv.Errorf("cache fail %v", err)
	}

}

func (c *Call) Many(path, query string, body []string, r *reply) {
	c.POST(path, query, body, r)
}

func (c *Call) Do(uri *url.URL, body []string) *reply {
	r := newReply(uri, body)
	n := len(body)
	path := fmt.Sprintf("/v1/security/%s", uri.Path)

	switch n {
	case 0:
		r.err = fmt.Errorf("not found body")
	case 1:
		c.One(uri, path, body, r)
	default:
		c.Many(path, uri.RawQuery, body, r)
	}

	return r
}
