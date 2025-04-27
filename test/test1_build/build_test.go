package test1_build

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"telego/test/testutil"
)

func TestBuild(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)

	// 复制或创建 compile_conf.yml
	confPath := filepath.Join(projectRoot, "compile_conf.yml")
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 创建新的配置文件
		confContent := `main_node:
  host: 127.0.0.1
  port: 2222
`
		if err := os.WriteFile(confPath, []byte(confContent), 0644); err != nil {
			t.Fatalf("创建配置文件失败: %v", err)
		}
	}

	// 使用 1.build.py 构建 telego
	cmd := exec.Command("python3", "1.build.py")
	cmd.Dir = projectRoot

	if err := testutil.RunCommand(t, cmd); err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	t.Log("构建成功")
}

func TestVerifyBinary(t *testing.T) {
	projectRoot := testutil.GetProjectRoot(t)
	binaryName := testutil.GetBinaryName()
	binaryPath := filepath.Join(projectRoot, "dist", binaryName)

	// 检查二进制文件是否存在
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("二进制文件不存在: %s", binaryPath)
	}

	// 执行 telego -v 检查版本
	cmd := exec.Command(binaryPath, "-v")
	cmd.Dir = projectRoot

	if err := testutil.RunCommand(t, cmd); err != nil {
		t.Fatalf("验证二进制文件失败: %v", err)
	}
	t.Log("二进制文件验证成功")
} 