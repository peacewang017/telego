package util

import (
	"fmt"
	"os"
	
	"github.com/fatih/color"
)

// GetPassword 尝试从环境变量获取密码，如果失败则提示用户输入
// 返回密码和获取是否成功的标志
func GetPassword() (string, bool) {
	// 先尝试从环境变量获取密码
	password, ok := os.LookupEnv("SSH_PW")
	
	// 如果环境变量中没有密码，提示用户输入
	if !ok {
		ok, password = StartTemporaryInputUI(
			color.GreenString("执行 sudo 命令需要输入密码"),
			"此处键入密码",
			"(回车确认，ctrl+c取消)")
		
		if !ok {
			return "", false
		}
	}
	
	return password, true
}