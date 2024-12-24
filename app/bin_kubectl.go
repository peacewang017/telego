package app

import "os/exec"

type BinManagerKubectl struct{}

func (k BinManagerKubectl) CheckInstalled() bool {
	// 尝试运行 "kubectl version --client" 命令以验证 kubectl 是否可用
	cmd := exec.Command("kubectl", "version", "--client")
	err := cmd.Run()
	if err != nil {
		// 如果命令执行失败，则认为 kubectl 未安装
		return false
	}
	// 如果命令成功执行，则认为 kubectl 已安装
	return true
}

func (k BinManagerKubectl) BinName() string {
	return "kubectl"
}

func (k BinManagerKubectl) SpecInstallFunc() func() error {
	return nil
}
