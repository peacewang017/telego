package testutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	SSH_PORT = 2222 // 非默认 SSH 端口
)

// BuildContext 构建上下文配置
type BuildContext struct {
	HOST_PROJECT_DIR string `yaml:"HOST_PROJECT_DIR"`
}

// GetHostProjectPath 获取主机项目目录路径
func GetHostProjectPath(t *testing.T) string {
	projectRoot := GetProjectRoot(t)
	buildContextPath := filepath.Join(projectRoot, "build_context.yml")

	// 读取 build_context.yml
	data, err := os.ReadFile(buildContextPath)
	if err != nil {
		t.Fatalf("读取 build_context.yml 失败: %v", err)
	}

	var buildContext BuildContext
	if err := yaml.Unmarshal(data, &buildContext); err != nil {
		t.Fatalf("解析 build_context.yml 失败: %v", err)
	}

	if buildContext.HOST_PROJECT_DIR == "" {
		t.Fatal("build_context.yml 中 HOST_PROJECT_DIR 为空")
	}

	return buildContext.HOST_PROJECT_DIR
}

// RunSSHDocker 运行一个用于 SSH 测试的 Docker 容器
func RunSSHDocker(t *testing.T) (string, func()) {
	// projectRoot := GetProjectRoot(t)
	hostProjectPath := GetHostProjectPath(t)

	// 拉取Python镜像并检查是否可用
	pullCmd := exec.Command("docker", "pull", "python:3.12.5")
	err := RunCommand(t, pullCmd)
	if err != nil {
		// 检查镜像是否已经存在
		checkCmd := exec.Command("docker", "image", "inspect", "python:3.12.5")
		if checkErr := checkCmd.Run(); checkErr != nil {
			// 镜像既不能拉取也不存在，测试无法继续
			t.Fatalf("拉取Python镜像失败，且本地无可用镜像: %v", err)
		}
		t.Logf("拉取Python镜像失败，但本地有可用镜像，继续测试: %v", err)
	}

	// 拉取并运行 Python 镜像，映射项目目录
	cmd := exec.Command("docker", "run", "-d",
		"-p", "2222:22",
		"-v", hostProjectPath+":/telego",
		"python:3.12.5",
		"tail", "-f", "/dev/null")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("启动 Python 容器失败: %v", err)
	}
	containerID := string(output[:12])

	// 等待容器就绪 - 鲁棒的方式
	maxAttempts := 50                       // 最多尝试50次
	waitInterval := 1000 * time.Millisecond // 每次等待1秒
	var ready bool

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pathCmd := exec.Command("docker", "exec", containerID, "echo", "Container is ready")
		if err := pathCmd.Run(); err == nil {
			ready = true
			t.Logf("容器就绪，第 %d 次尝试成功", attempt)
			break
		}
		t.Logf("等待容器就绪...尝试 %d/%d", attempt, maxAttempts)
		time.Sleep(waitInterval)
	}

	if !ready {
		t.Fatalf("容器启动超时，在 %d 次尝试后仍未就绪", maxAttempts)
	}

	// 检查 PATH 环境变量
	pathCmd := exec.Command("docker", "exec", containerID, "echo", "$PATH")
	if err := RunCommand(t, pathCmd); err != nil {
		t.Fatalf("检查 PATH 失败: %v", err)
	}

	// 在容器内执行拷贝命令
	copyCmd := exec.Command("docker", "exec", containerID,
		"cp", fmt.Sprintf("/telego/dist/%s", GetBinaryName()), "/usr/bin/telego")
	if err := RunCommand(t, copyCmd); err != nil {
		t.Fatalf("复制telego二进制文件失败: %v", err)
	}

	// 启用 SSH 密码认证
	sshCmd := exec.Command("docker", "exec", containerID,
		"telego", "ssh-passwd-auth", "--enable", "true")

	// 获取命令的输出
	stdout, err := sshCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("获取标准输出失败: %v", err)
	}
	stderr, err := sshCmd.StderrPipe()
	if err != nil {
		t.Fatalf("获取标准错误失败: %v", err)
	}

	// 启动命令
	if err := sshCmd.Start(); err != nil {
		t.Fatalf("启动命令失败: %v", err)
	}

	// 读取输出
	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	// 等待命令完成
	if err := sshCmd.Wait(); err != nil {
		t.Fatalf("启用 SSH 密码认证失败: %v\n标准输出: %s\n标准错误: %s",
			err, string(stdoutBytes), string(stderrBytes))
	}

	// 输出成功信息
	t.Logf("SSH 密码认证已启用\n标准输出: %s", string(stdoutBytes))

	// 在容器内执行 Python 脚本
	scriptCmd := exec.Command("docker", "exec", containerID,
		"python", "/telego/scripts/systemctl_docker.py")
	if err := RunCommand(t, scriptCmd); err != nil {
		t.Fatalf("执行 systemctl_docker.py 脚本失败: %v", err)
	}

	// 返回容器 ID 和清理函数
	return containerID, func() {
		exec.Command("docker", "rm", "-f", containerID).Run()
	}
}

// RunCommand 运行命令并实时输出
func RunCommand(t *testing.T, cmd *exec.Cmd) error {
	t.Logf("\n>>> 运行命令: %v", cmd.Args)
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
