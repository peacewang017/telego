package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// GetProjectRoot 获取项目根目录路径
func GetProjectRoot(t *testing.T) string {
	// 获取当前测试文件的路径
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法获取当前文件路径")
	}

	// 从当前目录开始向上遍历
	currentDir := filepath.Dir(filename)
	for {
		// 检查是否存在所需文件aok
		genMenuPath := filepath.Join(currentDir, "gen_menu.py")
		setupPath := filepath.Join(currentDir, "0.setup_build_and_run_shortcut.py")

		_, genMenuErr := os.Stat(genMenuPath)
		_, setupErr := os.Stat(setupPath)

		if genMenuErr == nil && setupErr == nil {
			t.Logf("找到项目根目录: %s", currentDir)
			return currentDir
		}

		// 获取父目录
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// 已经到达根目录
			t.Fatalf("无法找到项目根目录，未找到 gen_menu.py 和 0.setup_build_and_run_shortcut.py")
		}
		currentDir = parentDir
	}
}
