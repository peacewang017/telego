package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirLinkToolStruct 提供处理目录和链接的便捷方法
type DirLinkToolStruct struct {}

// 全局变量实例
var DirLinkTool = DirLinkToolStruct{}

// IsLinkDir 判断路径是否为指向目录的链接
func (d *DirLinkToolStruct) IsLinkDir(path string) (bool, error) {
	// 获取文件信息，不跟随符号链接
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 必须是符号链接
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	// 获取链接指向的目标路径
	dest, err := os.Readlink(path)
	if err != nil {
		return false, fmt.Errorf("读取链接目标失败: %w", err)
	}

	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(path), dest)
	}

	// 获取目标文件信息
	destInfo, err := os.Stat(dest)
	if err != nil {
		return false, fmt.Errorf("获取链接目标信息失败: %w", err)
	}

	// 判断目标是否为目录
	return destInfo.IsDir(), nil
}


// IsDir 判断路径是否为目录或指向目录的链接
func (d *DirLinkToolStruct) IsDirOrLinkDir(path string) (bool, error) {
	// 获取文件信息，不跟随符号链接
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 如果本身就是目录，直接返回true
	if info.IsDir() {
		return true, nil
	}

	// 如果是符号链接，检查目标是否为目录
	if info.Mode()&os.ModeSymlink != 0 {
		isLinkDir, err := d.IsLinkDir(path)
		if err != nil {
			return false, err
		}
		return isLinkDir, nil
	}

	// 既不是目录也不是链接
	return false, nil
}

// List 列出目录或链接目录的内容
// path: 目录路径或指向目录的链接
// 返回目录条目和可能的错误
func (d *DirLinkToolStruct) List(path string) ([]os.DirEntry, error) {
	isDir, err := d.IsDirOrLinkDir(path)
	if err != nil {
		return nil, fmt.Errorf("检查路径类型失败: %w", err)
	}

	if !isDir {
		return nil, fmt.Errorf("%s 不是目录或指向目录的链接", path)
	}

	// 获取文件信息，不跟随符号链接
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 如果是符号链接，获取指向的实际目录路径
	actualPath := path
	if info.Mode()&os.ModeSymlink != 0 {
		dest, err := os.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("读取链接目标失败: %w", err)
		}

		if !filepath.IsAbs(dest) {
			dest = filepath.Join(filepath.Dir(path), dest)
		}
		actualPath = dest
	}

	// 读取目录内容
	entries, err := os.ReadDir(actualPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录内容失败: %w", err)
	}

	return entries, nil
} 