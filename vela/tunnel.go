package vela

import (
	opcode "github.com/vela-ssoc/vela-kit/opcode"
	"io"
	"net/http"
)

type HTTPStream interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type HTTPResponse interface {
	Close()
	JSON(interface{}) error
	SaveFile(string) (string, error)
	Len() int64
	Read([]byte) (int, error)
	StatusCode() int
	H(string) string
	Header() http.Header
	E() error
}

type TnlByEnv interface { //tunnel by env
	OnConnect(name string, todo func() error)
	TnlName() string
	TnlVersion() string
	TnlIsDown() bool
	TnlSend(opcode.Opcode, interface{}) error
	DoHTTP(*http.Request) HTTPResponse
	HTTP(string, string, string, io.Reader, http.Header) HTTPResponse
	GET(string, string) HTTPResponse
	PostJSON(string, interface{}, interface{}) error
	Stream(string, interface{}) (HTTPStream, error)
}
