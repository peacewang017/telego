package test3_main_node_config

import (
	"os"
	"os/exec"
	"telego/test/testutil"
	"telego/util"
	"testing"
)

func TestSSHKeyGeneration(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)

	// 启动 SSH 测试容器
	_, cleanup := testutil.RunSSHDocker(t)
	defer cleanup()

	// 生成 SSH 密钥
	os.Setenv("SSH_PW", util.MainNodeUser)

	// // suppose to be failed
	// cmd := exec.Command("telego", "cmd", "--cmd", "/update_config/ssh_config/1.gen_or_get_key")
	// cmd.Dir = projectRoot

	// if err := testutil.RunCommand(t, cmd); err != nil {
	// 	t.Logf("生成 SSH 密钥失败，只有初始化fileserver后才会成功: %v", err)
	// } else {
	// 	t.Fatalf("生成 SSH 密钥成功，理论上未初始化main node file server，不应该成功")
	// }

	// init fileserver
	cmd := exec.Command("telego", "cmd", "--cmd", "/update_config/start_mainnode_fileserver")
	cmd.Dir = projectRoot
	if err := testutil.RunCommand(t, cmd); err != nil {
		t.Fatalf("初始化main node file server失败: %v", err)
	}

	// 重新生成 SSH 密钥
	cmd = exec.Command("telego", "cmd", "--cmd", "/update_config/ssh_config/1.gen_or_get_key")
	cmd.Dir = projectRoot
	if err := testutil.RunCommand(t, cmd); err != nil {
		t.Fatalf("重新生成 SSH 密钥失败: %v", err)
	}

	// 测试 SSH 连接
	sshCmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", "2222", util.MainNodeUser+"@"+util.MainNodeIp, "echo", "test")
	if err := testutil.RunCommand(t, sshCmd); err != nil {
		t.Fatalf("SSH 连接 abc 测试失败: %v", err)
	}

	t.Log("SSH 密钥生成和连接测试成功")
}
