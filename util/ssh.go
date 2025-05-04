package util

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/util/homedir"
)

// 扫描并加载所有可能的私钥文件
func findPrivateKeys(dir string) []string {
	var keys []string
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 过滤出可能的私钥文件
		if !d.IsDir() && (strings.Contains(d.Name(), "id_") && !strings.HasSuffix(d.Name(), ".pub")) {
			keys = append(keys, path)
		}
		return nil
	})
	return keys
}

// 尝试使用私钥连接服务器
func tryKey(server, username, keyPath string, port string) bool {
	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		fmt.Printf("无法读取私钥文件 %s: %v\n", keyPath, err)
		return false
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte{})
		if err != nil {
			fmt.Printf("无法解析私钥文件 %s: %v\n", keyPath, err)
			return false
		}
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 简化处理，请根据需要增强安全性
	}

	client, err := ssh.Dial("tcp", server+":"+port, config)
	if err != nil {
		Logger.Debugf("使用私钥 %s 连接失败: %v\n", keyPath, err)
		// fmt.Printf("使用私钥 %s 连接失败: %v\n", keyPath, err)
		return false
	}
	defer client.Close()

	// fmt.Printf("成功使用私钥 %s 连接到服务器！\n", keyPath)
	return true
}

func sshWithConf(server string, port string, config *ssh.ClientConfig) (*ssh.Client, *ssh.Session, error) {
	client, err := ssh.Dial("tcp", server+":"+port, config)
	if err != nil {
		// fmt.Printf("使用私钥 %s 连接服务器失败: %v\n", keyPath, err)
		return nil, nil, fmt.Errorf("连接服务器失败: %v\n", err)
	}
	// defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		// fmt.Printf("创建 SSH 会话失败: %v\n", err)
		return nil, nil, fmt.Errorf("创建 SSH 会话失败: %v\n", err)
	}

	return client, session, nil
}

func sshWithPasswd(server, username, passwd string, port string) (*ssh.Client, *ssh.Session, error) {
	Logger.Debugf("sshWithPasswd to %s@%s:%s", username, server, port)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	c, s, e := sshWithConf(server, port, config)
	if e != nil {
		return c, s, fmt.Errorf("使用密码连接失败：%v", e)
	} else {
		return c, s, nil
	}
}

// 使用找到的私钥连接并执行命令
func sshWithKey(server, username, keyPath string, port string) (*ssh.Client, *ssh.Session, error) {
	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		// fmt.Printf("无法读取私钥文件 %s: %v\n", keyPath, err)
		return nil, nil, fmt.Errorf("无法读取私钥文件 %s: %v\n", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		// fmt.Printf("无法解析私钥文件 %s: %v\n", keyPath, err)
		return nil, nil, fmt.Errorf("无法解析私钥文件 %s: %v\n", keyPath, err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	c, s, e := sshWithConf(server, port, config)
	if e != nil {
		return c, s, fmt.Errorf("使用私钥 %s 连接失败：%v", keyPath, e)
	} else {
		return c, s, nil
	}
}

// left 'usePasswd'
func sshSession(server, user, usePasswd string, specPort ...string) (*ssh.Client, *ssh.Session, error) {
	// 设置默认端口
	port := "22"

	// 如果提供了指定端口参数，则使用指定的端口
	if len(specPort) > 0 && specPort[0] != "" {
		port = specPort[0]
	}

	if usePasswd != "" {
		return sshWithPasswd(server, user, usePasswd, port)
	} else {
		// 扫描 ~/.ssh 目录下的所有私钥
		sshDir := filepath.Join(homedir.HomeDir(), ".ssh")
		// expanduser

		keys := findPrivateKeys(sshDir)
		if len(keys) == 0 {
			// fmt.Println("未找到任何私钥文件。")
			return nil, nil, fmt.Errorf("未在 %s 目录下找到任何私钥文件。", sshDir)
		}

		// 尝试每个私钥文件
		findKey := ""
		for _, key := range keys {
			// fmt.Printf("尝试使用私钥: %s\n", key)
			if tryKey(server, user, key, port) {
				findKey = key
				// return
				break
			}
		}

		if findKey == "" {
			// fmt.Println("未找到任何私钥文件。")
			return nil, nil, fmt.Errorf("未找到任何私钥文件。")
		}
		return sshWithKey(server, user, findKey, port)
	}

	// fmt.Println("未能使用任何私钥成功连接到服务器。")
}
