package app

import (
	"telego/util"
	"time"
)

type BinManagerWinfsp struct{}

func (k BinManagerWinfsp) CheckInstalled() bool {
	if !util.IsWindows() {
		return true
	}

	cmd, err := util.ModRunCmd.NewBuilder("rclone", "mount", "D:", "K:").AsyncRun()
	defer cmd.Process.Kill()
	if err != nil {
		return false
	}

	time.Sleep(2 * time.Second)
	if cmd.Err != nil {
		return false
	}

	// 如果命令成功执行，则认为 Winfsp 已安装
	return true
}

func (k BinManagerWinfsp) BinName() string {
	return "winfsp"
}

func (k BinManagerWinfsp) SpecInstallFunc() func() error {
	return nil
}
