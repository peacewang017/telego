package app

import "os/exec"

type BinManagerK3s struct{}

func (k BinManagerK3s) CheckInstalled() bool {
	// 尝试运行 "K3s version --client" 命令以验证 K3s 是否可用
	cmd := exec.Command("k3s", "--version")
	err := cmd.Run()
	if err != nil {
		// 如果命令执行失败，则认为 K3s 未安装
		return false
	}
	// 如果命令成功执行，则认为 K3s 已安装
	return true
}

func (k BinManagerK3s) BinName() string {
	return "k3s"
}

func (k BinManagerK3s) SpecInstallFunc() func() error {
	return nil
}
