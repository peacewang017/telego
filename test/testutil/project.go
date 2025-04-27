package testutil

import (
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
		// 检查是否存在所需文件
		genMenuPath := filepath.Join(currentDir, "gen_menu.py")
		setupPath := filepath.Join(currentDir, "0.setup_build_and_run_shortcut.py")
		
		genMenuExists := fileExists(genMenuPath)
		setupExists := fileExists(setupPath)
		
		if genMenuExists && setupExists {
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

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	_, err = filepath.Stat(path)
	return err == nil
} 