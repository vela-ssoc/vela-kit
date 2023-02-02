//go:build !windows && !plan9
// +build !windows,!plan9

package agent

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

type program struct {
	dir      string
	exe      string
	log      string
	cmd      exec.Cmd
	ss       service.Service
	fn       construct
	shutdown bool
}

func (p *program) Start(ss service.Service) error {
	go p.fork()
	return nil
}

func (p *program) fork() {
	runtime.LockOSThread()
	run(p.fn, &p.cmd, &p.shutdown)
}

func (p *program) Stop(s service.Service) error {
	p.shutdown = true
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func already(ss service.Service) {
	s, _ := ss.Status()
	if s == service.StatusUnknown {
		return
	}

	fmt.Printf("正在卸载服务\n")
	if e := ss.Uninstall(); e != nil {
		fmt.Printf("服务卸载失败%v\n", e)
	} else {
		fmt.Printf("服务卸载成功\n")
	}
}

func Exe() string {
	exe, err := os.Executable()
	if err == nil {
		return exe
	}

	exe, err = filepath.Abs(os.Args[0])
	if err == nil {
		return exe
	}

	return "/usr/local/ssoc/ssc"

}

func newP() *program {
	exe := Exe()
	p := &program{
		exe:      exe,
		shutdown: false,
		dir:      filepath.Dir(exe),
	}

	p.log = p.dir + "/logs/daemon.log"
	return p
}

func newSSC(fn construct) (service.Service, error) {
	p := newP()
	ssc := &service.Config{
		Name:             "ssc",
		DisplayName:      "SSOC Sensor",
		Description:      "EastMoney Security Management Platform",
		Arguments:        []string{"service"},
		Executable:       p.exe,
		WorkingDirectory: p.dir,
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target",
		},
	}

	ss, err := service.New(p, ssc)
	if err != nil {
		return nil, err
	}

	p.ss = ss
	p.fn = fn
	return ss, nil
}

func Install(fn construct) {
	ss, err := newSSC(fn)
	if err != nil {
		fmt.Printf("ssc sensor service error %v\n", err)
		return
	}

	fmt.Printf("正在安装服务\n")
	already(ss)

	if e := ss.Install(); e != nil {
		fmt.Printf("服务安装失败%s\n", e)
		return
	}
	fmt.Printf("服务安装成功\n")
}

func Uninstall(fn construct) {
	output, file := auxlib.Output()
	if file != nil {
		defer func() { _ = file.Close() }()
	}
	ss, err := newSSC(fn)
	if err != nil {
		output(`"msg":"ssc sensor service error %v"`, err)
		return
	}

	if e := ss.Uninstall(); e != nil {
		output(`"msg":"服务卸载失败%v"`, e)
		return
	}
	output(`"msg":"服务卸载成功"`)
}

func Service(fn construct) {
	output, file := auxlib.Output()
	if file != nil {
		defer func() { _ = file.Close() }()
	}

	ss, err := newSSC(fn)
	if err != nil {
		output(`"msg":"start ssc by service error %v"`, err)
		Start(fn)
		return
	}

	err = ss.Run()
	if err != nil {
		output(`"msg":"run ssc by service error %v"`, err)
		return
	}

	output(`"msg":"ssc service exit"`)
}

func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}

func executable(output func(string, ...interface{})) string {
	exe := Exe()
	if exe == "" {
		output(`"msg":"ssc executable got fail"`)
		return ""
	}

	if hi, e := binary.Decode(exe); e != nil {
		output(`"msg":"ssc %s binary decode error %v"`, exe, e)
		return ""
	} else {
		output(`"msg":"ssc %s binary code succeed"`, hi)
		return exe
	}
}
