package test3_main_node_config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"telego/util"
)

func TestMainNodeConfig(t *testing.T) {
	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		t.Fatalf("获取可执行文件路径失败: %v", err)
	}
	projectRoot := filepath.Dir(filepath.Dir(exePath))

	// 使用 1.build.py 构建 telego
	cmd := exec.Command("python3", "1.build.py")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("构建失败: %v\n输出: %s", err, string(output))
	}

	// 测试主节点配置
	t.Log("测试主节点配置...")
	
	// 测试读取配置
	reader := util.MainNodeConfReader{}
	
	// 测试读取 SSH 公钥
	pubkey, err := reader.ReadSecretConf(util.SecretConfTypeSshPublic{})
	if err != nil {
		t.Fatalf("读取 SSH 公钥失败: %v", err)
	}
	t.Log("SSH 公钥读取成功")

	// 测试读取 SSH 私钥
	privkey, err := reader.ReadSecretConf(util.SecretConfTypeSshPrivate{})
	if err != nil {
		t.Fatalf("读取 SSH 私钥失败: %v", err)
	}
	t.Log("SSH 私钥读取成功")

	// 测试写入配置
	writer := util.MainNodeConfWriter{}
	
	// 测试写入 SSH 公钥
	err = writer.WriteSecretConf(util.SecretConfTypeSshPublic{}, pubkey)
	if err != nil {
		t.Fatalf("写入 SSH 公钥失败: %v", err)
	}
	t.Log("SSH 公钥写入成功")

	// 测试写入 SSH 私钥
	err = writer.WriteSecretConf(util.SecretConfTypeSshPrivate{}, privkey)
	if err != nil {
		t.Fatalf("写入 SSH 私钥失败: %v", err)
	}
	t.Log("SSH 私钥写入成功")

	t.Log("主节点配置测试成功")
} 