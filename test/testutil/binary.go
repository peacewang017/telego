//go:build test

package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// GetBinaryName 获取当前平台的二进制文件名
func GetBinaryName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	binaryName := "telego_" + goos + "_" + goarch
	if goos == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

// CopyBinaryToSystem 将二进制文件复制到系统目录
func CopyBinaryToSystem(t *testing.T, projectRoot string) error {
	binaryName := GetBinaryName()
	sourcePath := filepath.Join(projectRoot, "dist", binaryName)
	targetPath := "/usr/bin/telego"

	// 读取源文件
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	// 写入目标文件
	return os.WriteFile(targetPath, data, 0755)
} 