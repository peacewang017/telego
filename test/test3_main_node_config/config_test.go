package test3_main_node_config

import (
	"os/exec"
	"testing"
	"telego/test/testutil"
)

func TestSSHKeyGeneration(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)

	// 启动 SSH 测试容器
	_, cleanup := testutil.RunSSHDocker(t)
	defer cleanup()

	// 生成 SSH 密钥
	cmd := exec.Command("telego", "cmd", "--cmd", "/update_config/ssh_config/1.gen_or_get_key")
	cmd.Dir = projectRoot

	if err := testutil.RunCommand(t, cmd); err != nil {
		t.Fatalf("生成 SSH 密钥失败: %v", err)
	}

	// 测试 SSH 连接
	sshCmd := exec.Command("ssh", "-p", "2222", "root@127.0.0.1", "echo", "test")
	if err := testutil.RunCommand(t, sshCmd); err != nil {
		t.Fatalf("SSH 连接测试失败: %v", err)
	}

	t.Log("SSH 密钥生成和连接测试成功")
}
