package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func ShellExec() string {
	maybeNames := []string{"bash", "sh", "fish", "ksh"}
	for _, name := range maybeNames {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return "shell-not-found"
	// // get which shell
	// shell := os.Getenv("SHELL")
	// if shell == "" {
	// 	shell = "/bin/bash"
	// }
	// return shell
}

// AdminUserConfig 定义管理员用户配置的结构
type AdminUserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// GetCurUserConfigPath 返回当前用户的配置文件路径
func GetCurUserConfigPath() string {
	configDir := "/teledeploy_secret/config"
	username := GetCurrentUser()
	return filepath.Join(configDir, "userconfig_"+username)
}

func GetUserConfigPath(username string) string {
	configDir := "/teledeploy_secret/config"
	return filepath.Join(configDir, "userconfig_"+username)
}

// ReadCurUserConfig 读取当前进程用户的配置文件
func ReadCurUserConfig() (*AdminUserConfig, error) {
	configPath := GetCurUserConfigPath()

	// 检查文件是否存在
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("user config file not found at %s", configPath)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading user config: %w", err)
	}

	// 解析YAML
	var config AdminUserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing user config: %w", err)
	}

	return &config, nil
}

// WriteAdminUserConfig 写入管理员用户配置到文件
// 可以指定要写入的用户名，默认写入adminuser
func WriteAdminUserConfig(config *AdminUserConfig) error {
	configPath := GetCurUserConfigPath()

	// 确保目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// 序列化配置为YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error serializing admin user config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing admin user config: %w", err)
	}

	return nil
}
