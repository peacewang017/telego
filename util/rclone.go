package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
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

type RcloneConfigType interface {
	RcloneConfigTypeDummyInterface()
}

type RcloneConfigTypeSsh struct{}

func (r RcloneConfigTypeSsh) RcloneConfigTypeDummyInterface() {
}

type RcloneConfiger struct {
	Type     RcloneConfigType
	Name     string
	Host     string
	Port     string
	User     string
	Password string
	Error    error
}

func NewRcloneConfiger(
	t RcloneConfigType,
	name string,
	address string,
) *RcloneConfiger {
	switch t := t.(type) {
	case RcloneConfigTypeSsh:
		// check address with format host:port or host
		if !func() bool {
			re := regexp.MustCompile(`^([a-zA-Z0-9.-]+)(?::\d+)?$`)
			return re.MatchString(address)
		}() {

			return &RcloneConfiger{
				Error: fmt.Errorf("不支持的 rclone ssh 地址: %+v", address),
			}
		}

		// Parse host and port
		hostport := strings.Split(address, ":")
		host := hostport[0]
		port := "22"
		if len(hostport) == 2 {
			port = hostport[1]
		}

		return &RcloneConfiger{
			Type: t,
			Host: host,
			Port: port,
			Name: name,
		}
	default:
		return &RcloneConfiger{
			Error: fmt.Errorf("不支持的配置类型: %+v", t),
		}
	}
}

func (r *RcloneConfiger) WithUser(user, pw string) *RcloneConfiger {
	fmt.Println(color.GreenString("rclone config with user %s", user))
	r.User = user
	r.Password = pw
	return r
}

func (r *RcloneConfiger) WithPort(port string) *RcloneConfiger {
	fmt.Println(color.GreenString("rclone config with port %s", port))
	r.Port = port
	return r
}

func (r *RcloneConfiger) DoConfig() error {
	if r.Error != nil {
		return r.Error
	}
	switch r.Type.(type) {
	case RcloneConfigTypeSsh:
		// rclone config
		// 配置项
		host := r.Host
		port := r.Port
		user := r.User

		encryptedPass := ""
		if r.Password != "" {
			encryptedPass_, err := ModRunCmd.NewBuilder("rclone", "obscure", r.Password).BlockRun()
			if err != nil {
				log.Fatalf("密码加密失败: %v\n", err)
			}
			encryptedPass = strings.ReplaceAll(strings.ReplaceAll(string(encryptedPass_), "\n", ""), " ", "")
		}
		// 使用 rclone obscure 加密密码

		// 输出加密的密码
		// fmt.Printf("加密后的密码: %s\n", encryptedPass)

		// 创建 rclone SFTP 远程节点
		cmds := []string{
			"rclone", "config", "create", r.Name, "sftp",
			"host=" + host,
			"user=" + user,
			"port=" + port,
			// "pass=" + encryptedPass,
			"use_insecure_cipher=false",
		}
		if encryptedPass != "" {
			cmds = append(cmds, "pass="+encryptedPass)
		}

		// fmt.Println(color.GreenString("rclone config create cmds: %+v", cmds))
		output, err := ModRunCmd.NewBuilder(cmds[0], cmds[1:]...).BlockRun()
		if err != nil {
			Logger.Errorf("rclone config create failed: %v with cmds: %v", err, cmds)
			return fmt.Errorf("远程节点 name:%s, host:%s, user:%s, port:%s, 配置失败: %s %v",
				r.Name, host, r.User, port, output, err)
		}
	default:
		return fmt.Errorf("不支持的远程节点类型: %+v", r.Type)
	}

	return nil
}

func RcloneDeleteRemote(name string) error {
	_, err := ModRunCmd.NewBuilder("rclone", "config", "delete", name).BlockRun()
	if err != nil {
		return fmt.Errorf("rclone 删除 %s 失败: %v", name, err)
	}
	return nil
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
	password, ok := GetPassword("使用rclone 远程访问需要配置密码")
	if !ok {
		fmt.Println("User canceled config rclone")
		os.Exit(1)
	}

	err = NewRcloneConfiger(RcloneConfigTypeSsh{}, MainNodeRcloneName, MainNodeIp).
		WithUser(MainNodeUser, password).
		WithPort(MainNodeSshPort).
		DoConfig()

	if err != nil {
		fmt.Println(color.RedString("配置Rclone失败: %v", err))
	} else {
		fmt.Printf("成功配置 SFTP 远程节点 %s。\n", MainNodeRcloneName)
	}
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

func RcloneSyncFileToFile(localPath string, remotePath string) error {
	// check local is file
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local file: %v", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("local path is a directory: %s", localPath)
	}

	// create temp file
	tempdir, err := os.MkdirTemp("", "rclone-sync-temp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempdir)

	_, err = ModRunCmd.NewBuilder("rclone", "copyto", localPath, remotePath).ShowProgress().BlockRun()
	if err != nil {
		return fmt.Errorf("failed to move temp file to remote path: %v", err)
	}

	return nil
}

func RcloneCheckDirExist(remotedir string) bool {
	_, err := ModRunCmd.NewBuilder("rclone", "lsd", remotedir).BlockRun()
	if err != nil {
		return false
	}
	return true
}
