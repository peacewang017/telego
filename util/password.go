package util

import (
	"os"

	"github.com/fatih/color"
)

// ReadAdminUserConfig 读取管理员用户配置
// func ReadAdminUserConfig() (*AdminUserConfig, error) {
// 	// TODO: 实现从配置文件读取管理员用户配置的逻辑
// 	return &AdminUserConfig{}, nil
// }

// GetPassword 尝试从环境变量获取密码，如果失败则提示用户输入
// 返回密码和获取是否成功的标志
func GetPassword(uiPrompt string) (string, bool) {
	// 1. 先尝试从环境变量获取密码
	password, ok := os.LookupEnv("SSH_PW")
	if ok {
		return password, true
	}

	// 2. 尝试从adminuser配置文件读取密码
	adminConfig, err := ReadCurUserConfig()
	if err == nil && adminConfig.Password != "" {
		Logger.Debugf("Using password from admin user config")
		return adminConfig.Password, true
	}

	// 3. 如果环境变量和配置文件中都没有密码，提示用户输入
	ok, password = StartTemporaryInputUI(
		color.GreenString(uiPrompt),
		"此处键入密码",
		"(回车确认，ctrl+c取消)")

	if !ok {
		return "", false
	}

	return password, true
}
