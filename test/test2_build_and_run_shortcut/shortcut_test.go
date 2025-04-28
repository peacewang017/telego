package test2_build_and_run_shortcut

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"telego/test/testutil"
)

func runCommand(t *testing.T, cmd *exec.Cmd) error {
	// 获取命令的标准输出和标准错误
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return err
	}

	// 创建读取器
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	// 实时读取并输出
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

	// 等待命令完成
	return cmd.Wait()
}

func getBinaryName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// 根据操作系统和架构生成二进制文件名
	binaryName := "telego_" + goos + "_" + goarch
	if goos == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

func TestSetupShortcut(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)

	// 使用 0.setup_build_and_run_shortcut.py 设置快捷方式
	cmd := exec.Command("python3", "0.setup_build_and_run_shortcut.py")
	cmd.Dir = projectRoot

	if err := runCommand(t, cmd); err != nil {
		t.Fatalf("设置快捷方式失败: %v", err)
	}
	t.Log("设置快捷方式成功")
}

func TestVerifyShortcut(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)
	binaryName := getBinaryName()
	binaryPath := filepath.Join(projectRoot, "dist", binaryName)

	// 检查二进制文件是否存在
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("二进制文件不存在: %s", binaryPath)
	}

	// 执行 telego -v 检查版本
	cmd := exec.Command(binaryPath, "-v")
	cmd.Dir = projectRoot

	if err := runCommand(t, cmd); err != nil {
		t.Fatalf("验证快捷方式失败: %v", err)
	}
	t.Log("快捷方式验证成功")
}
