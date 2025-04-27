package testutil

import (
	"bufio"
	"io"
	"os/exec"
	"testing"
	"time"
)

const (
	SSH_PORT = 2222  // 非默认 SSH 端口
)

// RunSSHDocker 运行一个用于 SSH 测试的 Docker 容器
func RunSSHDocker(t *testing.T) (string, func()) {
	// 拉取并运行 SSH 测试容器
	cmd := exec.Command("docker", "run", "-d", "-p", "2222:22", "linuxserver/openssh-server")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("启动 SSH 容器失败: %v", err)
	}
	containerID := string(output[:12])

	// 等待容器启动
	time.Sleep(5 * time.Second)

	// 返回容器 ID 和清理函数
	return containerID, func() {
		exec.Command("docker", "rm", "-f", containerID).Run()
	}
}

// RunCommand 运行命令并实时输出
func RunCommand(t *testing.T, cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	go func() {
		for {
			line, err := stdoutReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					t.Logf("读取标准输出错误: %v", err)
				}
				break
			}
			t.Log(line)
		}
	}()

	go func() {
		for {
			line, err := stderrReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					t.Logf("读取标准错误错误: %v", err)
				}
				break
			}
			t.Log(line)
		}
	}()

	return cmd.Wait()
} 