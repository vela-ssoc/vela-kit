package tunnel

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type HTTPResponse struct {
	res   *http.Response
	Error error
}

func (h HTTPResponse) E() error {
	return h.Error
}

func (h HTTPResponse) H(key string) string {
	if h.res == nil {
		return ""
	}
	return h.res.Header.Get(key)
}

func (h HTTPResponse) Header() http.Header {
	if h.res == nil {
		return nil
	}
	return h.res.Header
}

func (h HTTPResponse) StatusCode() int {
	if h.res != nil {
		return h.res.StatusCode
	}
	return http.StatusBadGateway
}

func (h HTTPResponse) Len() int64 {
	if h.res != nil {
		return h.res.ContentLength
	}

	return 0
}

func (h HTTPResponse) Close() {
	if h.res != nil {
		_ = h.res.Body.Close()
	}
}

func (h HTTPResponse) Read(buf []byte) (int, error) {
	if h.res != nil {
		return h.res.Body.Read(buf)
	}
	return 0, h.Error
}

func (h HTTPResponse) JSON(v interface{}) error {
	if h.Error != nil {
		return h.Error
	}
	defer func() { _ = h.res.Body.Close() }()

	if h.res.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(h.res.Body).Decode(v)
}

// SaveFile 将返回报文保存为文件并返回 md5
func (h HTTPResponse) SaveFile(path string) (string, error) {
	if h.Error != nil {
		return "", h.Error
	}
	defer func() { _ = h.res.Body.Close() }()

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, os.ModePerm)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	check := md5.New()
	if _, err = io.Copy(f, io.TeeReader(h.res.Body, check)); err != nil {
		return "", err
	}

	return hex.EncodeToString(check.Sum(nil)), nil
}

func (c Client) Do(r *http.Request) *HTTPResponse {
	if conn := c.conn; conn != nil {
		r.Header.Add(headerAuthorization, conn.Claim().Token)
	}

	res, err := c.client.Do(r)
	if err != nil {
		return &HTTPResponse{res: res, Error: err}
	}

	code := res.StatusCode
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(res.Body, buf)
		_ = res.Body.Close()
		he := &HTTPError{Code: code, Text: string(buf[:n])}
		return &HTTPResponse{Error: he}
	}

	return &HTTPResponse{res: res, Error: err}
}

func (c Client) HTTP(method, path, query string, body io.Reader, header http.Header) *HTTPResponse {
	dest := c.address.appendToHTTP(path, query)
	dest.RawQuery = query
	r, err := http.NewRequest(method, dest.String(), body)
	if err != nil {
		return nil
	}
	for k, v := range header {
		r.Header.Set(k, strings.Join(v, ", "))
	}
	return c.Do(r)
}

func (c Client) PostJSON(path string, data, reply interface{}) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}
	header := http.Header{"Content-Type": []string{"application/json"}}
	res := c.HTTP(http.MethodPost, path, "", buf, header)
	return res.JSON(reply)
}
