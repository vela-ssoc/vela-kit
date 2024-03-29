package powershell

import (
	"bytes"
	"context"
	"io"
	"os/exec"
)

// Powershell structure
type Powershell struct {
	handle *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// NewShell returns a new pointer to a Powershell structure
func NewShell() (p *Powershell, err error) {
	p = &Powershell{}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.handle = exec.CommandContext(p.ctx, "powershell", "-Command", "-")
	p.stdin, err = p.handle.StdinPipe()
	if err != nil {
		return
	}
	p.stdout, err = p.handle.StdoutPipe()
	if err != nil {
		return
	}
	p.stderr, err = p.handle.StderrPipe()
	if err != nil {
		return
	}
	// Start powershell command
	err = p.handle.Start()
	return
}

// execute from any kind of reader
func (p *Powershell) execute(r io.Reader) {
	b, _ := io.ReadAll(r)
	b = append(bytes.TrimRight(b, "\n"), '\n')
	p.stdin.Write(b)
}

// ExecuteString instruct the shell to execute the script provided as parameter
func (p *Powershell) ExecuteString(script string) {
	p.execute(bytes.NewReader([]byte(script)))
}

// HitReturn sends a new line to stdin
func (p *Powershell) HitReturn() {
	p.ExecuteString("\n")
}

// ImportFunction imports a function into running shell
func (p *Powershell) ImportFunction(code string) {
	p.ExecuteString(code)
	p.HitReturn()
}

// Exit exits the powershell console
func (p *Powershell) Exit() error {
	p.ExecuteString("Exit")
	return p.handle.Wait()
}

// Kill kills the current shell struct
func (p *Powershell) Kill() error {
	p.cancel()
	p.stdin.Close()
	p.stdout.Close()
	p.stderr.Close()
	return p.handle.Wait()
}
