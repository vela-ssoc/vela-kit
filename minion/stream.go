package minion

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"gopkg.in/tomb.v2"
	"io"
	"reflect"
	"time"
)

var (
	streamTypeOf = reflect.TypeOf((*stream)(nil)).String()
)

type stream struct {
	lua.SuperVelaData
	tx     Tx
	cfg    *config
	tom    *tomb.Tomb
	socket vela.HTTPStream
}

func newStream(cfg *config, tx Tx) *stream {
	st := &stream{tx: tx, cfg: cfg}
	st.V(lua.VTInit, fmt.Sprintf("%s.%s", streamTypeOf, tx.Type()))
	return st
}

func (st *stream) CodeVM() string {
	return st.tx.CodeVM()
}

func (st *stream) Name() string {
	return st.cfg.Name
}

func (st *stream) Type() string {
	return "stream." + st.tx.Type()
}

func (st *stream) ReloadConn() error {
	conn, err := xEnv.Stream(st.tx.Type(), st.cfg.Data)
	if err != nil {
		return err
	}

	st.socket.Close() //old socket
	st.socket = conn
	return nil
}

func (st *stream) Close() error {
	if st.IsRun() {
		st.tom.Kill(fmt.Errorf("close stream"))
		err := st.socket.Close()
		st.socket = nil
		return err
	}
	return fmt.Errorf("%s not running", st.Name())
}

func (st *stream) Start() error {
	conn, err := xEnv.Stream(st.tx.Type(), st.cfg.Data)
	if err != nil {
		return err
	}

	st.tom = new(tomb.Tomb)
	st.socket = conn
	return nil
}

func (st *stream) Read(buff []byte) (rn int, err error) {
	if !st.IsRun() || st.socket == nil {
		rn = 0
		err = fmt.Errorf("%s inactive", st.Name())
		return
	}
	return st.socket.Read(buff)
}

func (st *stream) write(data []byte) (wn int, err error) {
	wn, err = st.socket.Write(data)
	if err == nil {
		return
	}

	if err != nil {
		st.ReloadConn()
		return wn, err
	}

	//info := err.Error()
	//if strings.HasSuffix(info, "connection timed out") {
	//	return wn, st.ReloadConn()
	//}
	return wn, err
}

func (st *stream) Write(data []byte) (wn int, err error) {
	if !st.IsRun() || st.socket == nil {
		wn = 0
		err = fmt.Errorf("%s inactive", st.Name())
		return
	}

	if len(data) <= 0 {
		return 0, nil
	}

	if st.tx == nil {
		return st.write(data)
	}

	b := st.tx.Handle(data)
	if b == nil {
		return
	}
	defer buffer.Put(b)

	return st.write(b.Bytes())
}

func (st *stream) ReadFrom(rio io.Reader) (rn int64, err error) {

	buf := make([]byte, 4096)

	var n, n2 int
	var wn int64

	for {
		select {
		case <-st.tom.Dying():
			return

		default:
			n, err = rio.Read(buf)
			rn += int64(n)
			if err != nil {
				return
			}
			n2, err = st.Write(buf[0:n])
			if err != nil {
				return
			}
			wn += int64(n2)
		}

	}
}

func (st *stream) WriteTo(wio io.Writer) (wn int64, err error) {
	buff := make([]byte, 4096)
	var rn, n2 int

	for {
		select {
		case <-st.tom.Dying():
			return

		default:
			rn, err = st.socket.Read(buff)

			switch err {
			case nil:
				n2, err = wio.Write(buff[:rn])
				wn += int64(n2)
				if err != nil {
					return
				}

			case io.EOF:
				<-time.After(10 * time.Millisecond)
				break

			default:
				return
			}
		}
	}
}
