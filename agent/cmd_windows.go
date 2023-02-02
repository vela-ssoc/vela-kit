//go:build windows
// +build windows

package agent

import (
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/binary"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	name = "ssc"
)

type program struct {
	shutdown bool
	cmd      exec.Cmd
	fn       construct
	output   func(string, ...interface{})
}

func (p *program) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (sec bool, errno uint32) {
	const accepts = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}
	s <- svc.Status{State: svc.Running, Accepts: accepts}

	go run(p.fn, &p.cmd, &p.shutdown)

	defer func() {
		p.shutdown = true
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
	}()

	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			return
		default:
		}
		s <- svc.Status{State: svc.StopPending}
	}
	return
}

func Install(_ construct) {
	output, file := auxlib.Output()
	if file != nil {
		defer func() { _ = file.Close() }()
	}

	conn, err := mgr.Connect()
	if err != nil {
		output(`"msg":"connet windows service error %v"`, err)
		return
	}

	defer func() { _ = conn.Disconnect() }()

	if sc, erx := conn.OpenService(name); erx == nil {
		_ = sc.Close()
		return
	}

	exe, erx := os.Executable()
	if erx != nil {
		output(`"msg":"ssc fileapth got fail %v"`, erx)
		return
	}

	cfg := mgr.Config{
		DisplayName:      "SSOC Sensor",
		Description:      "EastMoney Security Management Platform",
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
	}

	ss, ers := conn.CreateService(name, exe, cfg, "service")
	if ers != nil {
		output(`"msg":"ssc create service error %v"`, ers)
		return
	}
	defer func() { _ = ss.Close() }()

	ras := []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 5 * time.Second}}

	if err = ss.SetRecoveryActions(ras, 0); err != nil {
		output(`"msg":"ssc create recovery action %v"`, err)
		return
	}

	eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	output(`"msg":"ssc install %s succeed"`, exe)
}

func Uninstall(_ construct) {
	cnn, _ := mgr.Connect()
	if cnn == nil {
		return
	}
	defer func() { _ = cnn.Disconnect() }()

	ss, _ := cnn.OpenService(name)
	if ss == nil {
		return
	}
	defer func() { _ = ss.Close() }()
	ss.Delete()
}

func Service(fn construct) {
	output, file := auxlib.Output()
	if file != nil {
		defer func() { _ = file.Close() }()
	}

	p := &program{fn: fn, shutdown: false, output: output}

	ok, err := svc.IsWindowsService()
	if err != nil {
		p.output("ssc service not windows %v\n", err)
		return
	}

	if !ok {
		return
	}

	err = svc.Run(name, p)
	if err == nil {
		p.output("ssc service exit error %v\n", err)
		return
	}

	p.output("ssc service exit\n")
	return

}

func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func exeWalk(current string) []string {
	var ret []string
	var mask []fs.FileInfo

	filepath.Walk(current, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if filepath.Ext(path) != ".exe" ||
			!strings.HasPrefix(info.Name(), "ssc-") ||
			filepath.Dir(path) != current {
			return nil
		}

		cAttr := info.Sys().(*syscall.Win32FileAttributeData)
		for i, stat := range mask {
			fAttr := stat.Sys().(*syscall.Win32FileAttributeData)
			if fAttr.CreationTime.Nanoseconds() < cAttr.CreationTime.Nanoseconds() {
				e := append([]string{}, ret[i:]...)
				s := append(ret[0:i], path)
				ret = append(s, e...)

				em := append([]os.FileInfo{}, mask[i:]...)
				sm := append(mask[0:i], info)
				mask = append(sm, em...)
				return nil
			}
		}

		ret = append(ret, path)
		mask = append(mask, info)
		return nil
	})

	return ret
}

func executable(output func(string, ...interface{})) string {
	exe, err := os.Executable()
	if err != nil {
		output(`"msg":"ssc executable got fail %v"`, err)
		return ""
	}
	current := filepath.Dir(exe)
	files := exeWalk(current)

	if len(files) == 0 {
		output(`"msg":"not found ssc file"`)
		return ""
	}

	for _, path := range files {
		hi, e := binary.Decode(path)
		if e == nil {
			output(`"msg":"ssc %s binary code succeed %+v"`, hi)
			return path
		}
		output(`"msg":"ssc %s binary decode error %v"`, path, e)
	}
	output(`"msg":"%+v not found valid ssc exe"`, files)
	return ""
}

func Exe() string {
	exe, _ := os.Executable()
	return exe
}
