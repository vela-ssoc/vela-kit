package dispatch

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/minion/tunnel"
	"github.com/vela-ssoc/vela-kit/opcode"
	"github.com/vela-ssoc/vela-kit/vela"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type message struct {
	cli *tunnel.Client
	rec *tunnel.Receive
}

type dispatch struct {
	xEnv         vela.Environment
	task         velaTask
	third        *thirdManager
	pmu          sync.RWMutex
	processes    map[opcode.Opcode]*process
	taskSyncing  int32
	thirdSyncing int32
	messages     chan *message
	upgrading    int32
}

func WithEnv(env vela.Environment) *dispatch {
	processes := make(map[opcode.Opcode]*process, 16)
	third := newThirdManager(env)

	messages := make(chan *message, 64)

	d := &dispatch{xEnv: env, task: velaTask{xEnv: env}, third: third, processes: processes, messages: messages}
	_ = d.register(opcode.OpSubstance, d.syncTask)
	_ = d.register(opcode.OpThird, d.syncThird)
	_ = d.register(opcode.OpReload, d.reloadSubstance)
	_ = d.register(opcode.OpOffline, d.opOffline)
	_ = d.register(opcode.OpDeleted, d.opDeleted)
	_ = d.register(opcode.OpUpgrade, d.opUpgrade)
	_ = d.register(opcode.OpResync, d.opSync)
	_ = d.register(opcode.OpThirdChange, d.OnThirdSync)

	// 执行收到的消息
	d.worker(16)

	return d
}

func (d *dispatch) OnConnect(cli *tunnel.Client) {
	_ = d.syncThird(cli)
	_ = d.syncTask(cli)
}

func (d *dispatch) OnMessage(cli *tunnel.Client, rec *tunnel.Receive) {
	d.messages <- &message{cli: cli, rec: rec}
}

func (d *dispatch) OnDisconnect(_ *tunnel.Client) {
}

func (d *dispatch) syncTask(cli *tunnel.Client) error {
	if !atomic.CompareAndSwapInt32(&d.taskSyncing, 0, 1) {
		return nil
	}
	defer atomic.CompareAndSwapInt32(&d.taskSyncing, 1, 0)

	d.task.sync(cli)

	return nil
}

func (d *dispatch) worker(n int) {
	for i := 0; i < n; i++ {
		go d.work()
	}
}

func (d *dispatch) work() {
	for msg := range d.messages {
		d.process(msg)
	}
}

func (d *dispatch) process(msg *message) {
	cli, rec := msg.cli, msg.rec
	opcode := rec.Opcode()
	d.xEnv.Warnf("执行命令: %s", opcode)
	d.pmu.RLock()
	proc := d.processes[opcode]
	d.pmu.RUnlock()
	if proc == nil {
		d.xEnv.Warnf("没有相关命令 process: %s", opcode)
		return
	}

	if err := proc.execute(cli, rec); err != nil {
		d.xEnv.Warnf("%s 处理发生错误: %v", opcode, err)
	} else {
		d.xEnv.Infof("%s 处理完毕", opcode)
	}
}

func (d *dispatch) syncThird(cli *tunnel.Client) error {
	if !atomic.CompareAndSwapInt32(&d.thirdSyncing, 0, 1) {
		return nil
	}
	defer atomic.CompareAndSwapInt32(&d.thirdSyncing, 1, 0)

	d.xEnv.Infof("收到 3rd 文件变动信令")
	d.third.sync(cli)
	return nil
}

func (d *dispatch) reloadSubstance(cli *tunnel.Client, dat *substance) error {
	return d.task.reload(cli, dat)
}

// opDeleted
func (d *dispatch) opDeleted(_ *tunnel.Client) error {
	d.xEnv.Warnf("节点被删除，理解退出程序")
	os.Exit(0)
	return nil
}

// opOffline
func (d *dispatch) opOffline(cli *tunnel.Client) error {
	d.xEnv.Warnf("节点下线")
	_ = cli.Close()
	os.Exit(0)
	return nil
}

type upgradeVO struct {
	Edition string `json:"edition"`
}

func (d *dispatch) request(cli *tunnel.Client, save, query string) error {
	r := cli.HTTP(http.MethodGet, "/v1/edition/upgrade",
		query, nil, nil)

	if r.Error != nil {
		return r.Error
	}

	switch r.StatusCode() {

	case http.StatusNoContent: //结束请求 不重启
		d.xEnv.Error("无需升级,已经是最新版本")
		return nil

	case http.StatusOK:
		_, err := r.SaveFile(save)
		return err

	case http.StatusNotModified:
		d.xEnv.Error("无需升级,已经是最新版本")
		return nil

	default:
		return fmt.Errorf("升级请求获取失败code:%d", r.StatusCode())

	}
}

func (d *dispatch) Download(cli *tunnel.Client, save, query string) error {
	err := d.request(cli, save, query)
	if err == nil {
		return nil
	}

	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()

	failure := 0
	for range tk.C {
		failure++
		if failure > 360 {
			return fmt.Errorf("升级包下载超过360次失败")
		}

		if err = d.request(cli, save, query); err == nil {
			return nil
		}
	}

	return fmt.Errorf("下载升级包失败")
}

// OpUpgrade
func (d *dispatch) opUpgrade(cli *tunnel.Client, vo *upgradeVO) error {
	if !atomic.CompareAndSwapInt32(&d.upgrading, 0, 1) {
		d.xEnv.Errorf("多次指令接收,正在升级中...")
		return nil
	}
	defer atomic.StoreInt32(&d.upgrading, 0)

	d.xEnv.Infof("节点升级到: %s", vo.Edition)

	// 获取当前文件的绝对路径
	abs, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}

	// 获取当前的工作目录
	workdir, name := filepath.Split(abs)
	ext := filepath.Ext(name)
	if len(ext) > 0 {
		name = strings.SplitN(name, ext, 2)[0]
	}

	backDir := filepath.Join(workdir, "backup")
	backName := filepath.Join(backDir, fmt.Sprintf("%s-%s%s", name, cli.Version(), ext))

	_ = os.RemoveAll(backDir) // 只备份本次的二进制包, 历史备份二进制包不留存, 简单粗暴: 删除备份目录, 将本次二进制放到备份目录
	if err = os.MkdirAll(backDir, os.ModePerm); err != nil {
		d.xEnv.Errorf("创建备份文件夹%s错误: %v", backDir, err)
		//return err
	}

	d.xEnv.Infof("开始备份当前二进制文件: %s ---> %s", abs, backName)
	cf, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer func() { _ = cf.Close() }()
	bf, err := os.OpenFile(backName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err == nil {
		defer func() { _ = bf.Close() }()
	} else {
		d.xEnv.Errorf("打开当前二进制 %s ---> %s 错误: %v", abs, backName, err)
	}

	if _, err = io.Copy(bf, cf); err != nil {
		d.xEnv.Errorf("备份当前二进制 %s ---> %s 错误: %v", abs, backName, err)
		//return err
	}

	// 下载最新版本
	save := filepath.Join(workdir, fmt.Sprintf("ssc-%d%s", time.Now().Unix(), ext))
	query := "edition=" + vo.Edition
	err = d.Download(cli, save, query)
	if err != nil {
		d.xEnv.Errorf("下载升级包失败%v", err)
		return err
	}

	_, err = os.Stat(save)
	if err != nil {
		fmt.Errorf("可执行程序未成功保存%v", err)
		return nil
	}

	switch runtime.GOOS {
	case "linux":
		// 刚刚下载的文件覆盖掉运行的文件名
		if err = os.RemoveAll(abs); err != nil {
			d.xEnv.Errorf("删除文件%s错误: %v", abs, err)
		}
		if err = os.Rename(save, abs); err != nil {
			d.xEnv.Errorf("升级包 %s -> %s 覆盖失败: %v", save, abs, err)
			return err
		}
	}

	os.Exit(0)

	return nil
}
