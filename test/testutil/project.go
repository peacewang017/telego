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

	// 获取项目根目录（向上导航两级）
	projectRoot := filepath.Dir(filepath.Dir(filename))
	t.Logf("项目根目录: %s", projectRoot)
	return projectRoot
} 