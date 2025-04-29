package util

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"encoding/base64"

	"github.com/fatih/color"
	"k8s.io/client-go/util/homedir"
)

// SystemType 系统类型接口
type SystemType interface {
	GetArchCmd() string
}

// LinuxSystem Linux 系统类型
type LinuxSystem struct{}

func (s LinuxSystem) GetArchCmd() string {
	return "uname -m"
}

// WindowsSystem Windows 系统类型
type WindowsSystem struct{}

func (s WindowsSystem) GetArchCmd() string {
	return "echo %PROCESSOR_ARCHITECTURE%"
}

// DarwinSystem macOS 系统类型
type DarwinSystem struct{}

func (s DarwinSystem) GetArchCmd() string {
	return "uname -m"
}

type UnknownSystem struct{}

func (s UnknownSystem) GetArchCmd() string {
	return ""
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func scanDirWithFilter(root string, filter func(entry os.DirEntry) bool) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var results []string
	for _, entry := range entries {
		if filter(entry) { // 根据回调函数进行过滤
			results = append(results, filepath.Join(root, entry.Name()))
		}
	}

	return results, nil
}

var entryDir = ""

func GetEntryDir() string {
	return entryDir
}

func SaveEntryDir() string {
	entryDir0, err := filepath.Abs(CurDir())
	if err != nil {
		entryDir = homedir.HomeDir()
	} else {
		entryDir = entryDir0
	}
	return entryDir
}

func CurDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		fmt.Println(color.RedString("get current dir failed %s", err))
		os.Exit(1)
	}
	return curDir
}

type CachedHasNetwork struct {
	v bool
}

var cachedHasNetwork *CachedHasNetwork = nil

const ArchArm64 = "arm64"
const ArchAmd64 = "amd64"

func GetCurrentArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "arm64", "aarch64":
		return ArchArm64
	case "amd64", "x86_64":
		return ArchAmd64
	default:
		return ArchAmd64 // Default to amd64 if unknown
	}
}

func HasNetwork() bool {
	if cachedHasNetwork != nil {
		return cachedHasNetwork.v
	}

	// 检查网络连接
	client := &http.Client{
		Timeout: 5 * time.Second, // 设置超时时间
	}
	resp, err := client.Get("https://www.baidu.com")
	if err != nil {
		cachedHasNetwork = &CachedHasNetwork{v: false}
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		cachedHasNetwork = &CachedHasNetwork{v: true}
		return true
	}

	cachedHasNetwork = &CachedHasNetwork{v: false}
	return false
}

func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error retrieving user:", err)
		return false
	}

	return currentUser.Uid == "0"
}

// CurrentTimeString returns the current time as a formatted string
// suitable for filenames (YYYY-MM-DD-HHMMSS format)
func CurrentTimeString() string {
	return time.Now().Format("2006-01-02-150405")
}

// IsLinux returns true if the current OS is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMacOS returns true if the current OS is macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// WriteFileWithContent writes content to a file with root privileges if needed
func WriteFileWithContent(path string, content string) (string, error) {
	// Encode content to base64
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))

	// Use telego command to execute the job
	output, err := ModRunCmd.RequireRootRunCmd("telego", "decode-base64-to-file",
		"--base64", encodedContent,
		"--path", path,
		"--mode", "0644")
	return output, err
}

// GetCurrentUser 获取当前进程的用户名
func GetCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		Logger.Warnf("Error retrieving current user: %v", err)
		return "adminuser"
	}
	return currentUser.Username
}
