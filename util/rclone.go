package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func isRcloneRemoteConfigured(remoteName string) (bool, error) {
	// 调用 `rclone listremotes` 获取已配置的 remotes
	cmd := exec.Command("rclone", "listremotes")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running rclone: %w", err)
	}

	// 获取命令输出
	output := out.String()

	// 检查输出中是否包含指定的 remote 节点
	remotes := strings.Split(output, "\n")
	for _, remote := range remotes {
		// 忽略空行并检查 remoteName
		if strings.TrimSpace(remote) == remoteName+":" {
			return true, nil
		}
	}

	// 如果找不到指定的 remote 节点
	return false, nil
}

func ConfigMainNodeRcloneIfNeed() {
	conf, err := isRcloneRemoteConfigured(MainNodeRcloneName)
	if err != nil {
		fmt.Println(color.RedString("isRcloneRemoteConfigured Error: %v\n", err))
		os.Exit(1)
	}
	if conf {
		return
	}

	// get environment SSH_PW
	password, ok := os.LookupEnv("SSH_PW")

	if !ok {
		ok, password = StartTemporaryInputUI(color.GreenString("上传teledeploy/teledeploy_secret需要配置main_node rclone, 请输入 %s:%s 密码:", MainNodeUser, MainNodeIp),
			"此处键入密码",
			"(回车确认，ctrl+c取消)")
		if !ok {
			fmt.Println("User canceled config rclone")
			os.Exit(1)
		}
	}

	// rclone config
	// 配置项
	remoteName := MainNodeRcloneName
	host := MainNodeIp
	user := MainNodeUser
	// password := "your_plain_password"
	// keyFile := "/path/to/your/private_key"
	port := "22"

	// 使用 rclone obscure 加密密码
	encryptedPass, err := ModRunCmd.NewBuilder("rclone", "obscure", password).BlockRun()
	if err != nil {
		log.Fatalf("密码加密失败: %v\n", err)
	}
	encryptedPass = strings.ReplaceAll(strings.ReplaceAll(string(encryptedPass), "\n", ""), " ", "")

	// 输出加密的密码
	// fmt.Printf("加密后的密码: %s\n", encryptedPass)

	// 创建 rclone SFTP 远程节点
	_, err = ModRunCmd.NewBuilder("rclone", "config", "create", remoteName, "sftp",
		"host", host,
		"user", user,
		"port", port,
		"pass", encryptedPass,
		// "key_file", keyFile,
		"use_insecure_cipher", "false",
	).BlockRun()
	if err != nil {
		log.Fatalf("远程节点配置失败: %v\n", err)
	}

	// fmt.Printf("成功配置 SFTP 远程节点 %s。\n", remoteName)
}

// don't do path check here
func RcloneSyncDirOrFileToDir(localPath string, remotePath string) error {
	// rclone sync -P {localPath} remote:{remotePath}
	_, err := ModRunCmd.ShowProgress("rclone", "sync", "-P", localPath, remotePath).BlockRun()
	if err != nil {
		fmt.Println(color.RedString("rclone sync failed: %v", err))
	}

	return err
}
