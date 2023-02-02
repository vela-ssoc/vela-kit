package env

import (
	"github.com/vela-ssoc/vela-kit/node"
	"net"
)

type broker struct {
	arch   string
	mac    net.HardwareAddr
	inet   net.IP
	inet6  net.IP
	remote net.Addr
	edit   string
}

func (env *Environment) Prefix() string {
	return node.Prefix()
}

func (env *Environment) Broker() string {
	if env.bkr.remote == nil {
		return ""
	}

	cnn, ok := env.bkr.remote.(*net.TCPAddr)
	if !ok {
		return ""
	}
	return cnn.IP.String()
}

func (env *Environment) ID() string {
	return node.ID()
}

func (env *Environment) Arch() string {
	return env.bkr.arch
}

func (env *Environment) Mac() string {
	return env.bkr.mac.String()
}

func (env *Environment) Inet() string {
	if env.bkr.inet == nil {
		return ""
	}

	return env.bkr.inet.String()
}

func (env *Environment) Inet6() string {
	if env.bkr.inet6 == nil {
		return ""
	}

	return env.bkr.inet6.String()
}

func (env *Environment) Edition() string {
	return env.bkr.edit
}

func (env *Environment) LocalAddr() string {
	return env.bkr.inet.String()
}

func (env *Environment) WithBroker(arch string, mac net.HardwareAddr, inet net.IP, inet6 net.IP, edit string, remote net.Addr) {
	env.bkr.arch = arch
	env.bkr.mac = mac
	env.bkr.inet = inet
	env.bkr.inet6 = inet6
	env.bkr.edit = edit
	env.bkr.remote = remote
}
