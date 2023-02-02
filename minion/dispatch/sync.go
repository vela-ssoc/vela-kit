package dispatch

import (
	"github.com/vela-ssoc/vela-kit/minion/tunnel"
	"github.com/vela-ssoc/vela-kit/safecall"
	"github.com/vela-ssoc/vela-kit/vela"
	"time"
)

func (d *dispatch) OnThirdSync(cli *tunnel.Client, v struct {
	Name string `json:"name"`
	Drop bool   `json:"drop"`
}) error {
	d.xEnv.OnThirdSync(v.Name, v.Drop)
	return nil
}

func (d *dispatch) opSync(cli *tunnel.Client) error {
	onE := func(err error) {
		d.xEnv.Errorf("sync task fail %v", err)
	}

	do := func() error {
		return d.xEnv.WakeupTask(vela.TRANSPORT)
	}

	safecall.New(true).Timeout(time.Minute).OnError(onE).Exec(do)
	d.task.sync(cli)
	return nil
}
