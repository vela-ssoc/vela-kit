package kind

import (
	"context"
	"fmt"
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/execpt"
	"net"
	"sync/atomic"
)

type logger interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Infof(string, ...interface{})
}

type Accept func(context.Context, net.Conn) error

type Listener struct {
	done uint32
	xEnv vela.Environment
	bind auxlib.URL
	fd   []net.Listener
	ch   chan net.Conn
	ctx  context.Context
	stop context.CancelFunc
}

func (ln *Listener) CloseActiveConn() {
	ln.stop()
	ln.ctx, ln.stop = context.WithCancel(context.Background())
}

func (ln *Listener) Done() bool {
	return atomic.LoadUint32(&ln.done) == 1
}

func (ln *Listener) shutdown() {
	atomic.StoreUint32(&ln.done, 1)
}

func (ln *Listener) newConn(accept Accept, conn net.Conn) {
	if e := accept(ln.ctx, conn); e != nil {
		ln.xEnv.Errorf("%s listen handler failure , error %v", conn.LocalAddr().String(), e)
	} else {
		ln.xEnv.Errorf("%s listen handler over", conn.LocalAddr().String())
	}
}

func (ln *Listener) loop(accept Accept, sock net.Listener) {
	defer sock.Close()

	for {
		conn, err := sock.Accept()
		if err == nil {
			go ln.newConn(accept, conn)
			continue
		}

		if ln.Done() {
			ln.xEnv.Errorf("%s listen done.", sock.Addr().String())
			return
		}

		ln.xEnv.Errorf("%s listen accpet fail %v", sock.Addr().String(), err)
		return
	}

}

func (ln *Listener) multipleH(accept Accept) error {
	for _, sock := range ln.fd {
		go ln.loop(accept, sock)
	}
	<-ln.ctx.Done()
	ln.xEnv.Errorf("%s multiple handle exit", ln.bind.String())
	return nil
}

func (ln *Listener) OnAccept(accept Accept) error {

	n := len(ln.fd)
	if n < 1 {
		return fmt.Errorf("not found ative listen fd")
	}

	if n == 1 {
		go ln.loop(accept, ln.fd[0])
		return nil
	} else {
		return ln.multipleH(accept)
	}
}

func (ln *Listener) Close() error {
	if ln == nil {
		return nil
	}

	ln.stop()
	ln.shutdown()
	me := execpt.New()
	for _, fd := range ln.fd {
		me.Try(fd.Addr().String(), fd.Close())
	}
	ln.fd = nil
	return me.Wrap()
}

func (ln *Listener) single() error {
	fd, err := net.Listen(ln.bind.Scheme(), ln.bind.Host())
	if err != nil {
		return err
	}
	ln.fd = []net.Listener{fd}
	return nil
}

// multiple tcp://192.168.0.1/?port=1024,65535&exclude=1,2,3,4
func (ln *Listener) multiple() error {
	ps := ln.bind.Ports()
	n := len(ps)
	if n == 0 {
		return fmt.Errorf("%s not found listen", ln.bind.String())
	}

	for i := 0; i < n; i++ {
		port := ps[i]
		fd, e := net.Listen(ln.bind.Scheme(), fmt.Sprintf("%s:%d", ln.bind.Hostname(), port))
		if e != nil {
			return fmt.Errorf("listen %s://%s:%d error %v", ln.bind.Scheme(), ln.bind.Hostname(), port, e)
			continue
		}
		ln.fd = append(ln.fd, fd)
	}

	if len(ln.fd) == 0 {
		return fmt.Errorf("%s listen fail", ln.bind.String())
	}

	return nil
}

func (ln *Listener) Start() error {
	if ln.bind.Port() != 0 {
		return ln.single()
	}

	return ln.multiple()
}

func Listen(env vela.Environment, bind auxlib.URL) (*Listener, error) {
	ctx, stop := context.WithCancel(context.Background())
	ln := &Listener{
		bind: bind,
		ctx:  ctx,
		stop: stop,
		done: 0,
		xEnv: env,
	}

	return ln, ln.Start()
}
