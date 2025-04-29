package util

// package main

// import "telego/app"

// func main() {
// 	app.Main()
// }

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
	"encoding/base64"
)

type NodeState struct {
	Host       string
	Output     string
	IsComplete bool
}

type RemoteControlModel struct {
	Nodes []NodeState
}

type NodeMsg struct {
	Index    int
	Output   string
	Complete bool
}

const (
	DonePrefix = "Done with "
)

func (m RemoteControlModel) Init() tea.Cmd {
	return nil
}

func (m RemoteControlModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case NodeMsg:
		// 更新节点状态
		m.Nodes[msg.Index].Output = msg.Output
		m.Nodes[msg.Index].IsComplete = msg.Complete

		// 检查所有节点是否完成
		allComplete := true
		for _, node := range m.Nodes {
			if !node.IsComplete {
				allComplete = false
				break
			}
		}

		// 如果所有节点完成，退出程序
		if allComplete {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m RemoteControlModel) View() string {
	view := color.BlueString("Remote Command Execution:\n\n")
	for _, node := range m.Nodes {
		status := color.BlueString("Running")
		if node.IsComplete {
			status = color.GreenString("Completed")
		}
		view += fmt.Sprintf("[%s] %s: %s\n", status, node.Host, node.Output)
	}
	return view
}



// GetRemoteArch 获取远程主机的架构信息
// hosts: 远程主机列表
// usePasswd: 密码
// currentSystems: 每个主机对应的系统类型
func GetRemoteArch(hosts []string, usePasswd string, currentSystems []SystemType) []string {
	if len(hosts) != len(currentSystems) {
		Logger.Errorf("hosts and systems length mismatch: hosts=%d, systems=%d", len(hosts), len(currentSystems))
		return make([]string, len(hosts))
	}

	// 为每个主机执行对应的命令
	results := make([]string, len(hosts))
	for i, host := range hosts {
		cmd := currentSystems[i].GetArchCmd()
		hostResults := StartRemoteCmds([]string{host}, cmd, usePasswd)
		if len(hostResults) > 0 {
			// 清理结果
			result := strings.ToLower(strings.TrimSpace(hostResults[0]))
			result = strings.ReplaceAll(result, "\n", "")
			result = strings.ReplaceAll(result, "\r", "")
			result = strings.ReplaceAll(result, " ", "")

			// 判断架构
			if result == "aarch64" || result == "arm64" || result == "arm64e" {
				results[i] = "arm64"
			} else if result == "x86_64" || result == "amd64" || result == "x64" {
				results[i] = "amd64"
			} else {
				results[i] = "unknown"
			}
		} else {
			results[i] = "unknown"
		}
	}

	Logger.Debugf("GetRemoteArch: %v", results)
	return results
}

// GetRemoteSys 获取远程主机的系统类型
// hosts: 远程主机列表
// usePasswd: 密码
func GetRemoteSys(hosts []string, usePasswd string) []SystemType {
	// 使用 uname 命令获取系统信息
	results := StartRemoteCmds(hosts, "uname -s", usePasswd)
	
	fmt.Println(color.BlueString("GetRemoteSys raw output: %v", results))
	Logger.Debugf("GetRemoteSys: %v", results)
	return funk.Map(results, func(result string) SystemType {
		// 清理结果
		result = strings.ToLower(strings.TrimSpace(result))
		result = strings.ReplaceAll(result, "\n", "")
		result = strings.ReplaceAll(result, "\r", "")
		result = strings.ReplaceAll(result, " ", "")

		// 判断系统类型
		switch result {
		case "linux":
			return LinuxSystem{}
		case "darwin":
			return DarwinSystem{}
		case "windows":
			return WindowsSystem{}
		default:
			// 如果 uname 命令失败，尝试使用其他方法
			// 检查是否存在 Windows 特有的环境变量
			winCheck := StartRemoteCmds(hosts, "echo %OS%", usePasswd)
			if len(winCheck) > 0 && strings.Contains(strings.ToLower(winCheck[0]), "windows") {
				return WindowsSystem{}
			}
			// 默认返回 Linux
			return UnknownSystem{}
		}
	}).([]SystemType)
}

// 为每个用户创建配置文件
func createUserConfigs(hosts []string, usePasswd string) error {
	// 使用 map 对用户进行去重
	uniqueUsers := make(map[string]bool)
	for _, host := range hosts {
		hostsplit := strings.Split(host, "@")
		if len(hostsplit) != 2 {
			return fmt.Errorf("invalid host format: %s", host)
		}
		uniqueUsers[hostsplit[0]] = true
	}

	// 为每个唯一用户创建配置文件
	for user := range uniqueUsers {
		// 创建用户特定的配置文件
		userConfig := AdminUserConfig{
			Username: user,
			Password: usePasswd,
		}

		// 创建用户特定的配置文件路径，添加 userconfig_ 前缀
		userConfigPath := fmt.Sprintf("/teledeploy_secret/config/userconfig_%s", user)
		
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(userConfigPath), 0755); err != nil {
			return fmt.Errorf("error creating config directory for user %s: %w", user, err)
		}

		// 序列化配置为YAML
		data, err := yaml.Marshal(userConfig)
		if err != nil {
			return fmt.Errorf("error serializing config for user %s: %w", user, err)
		}

		// 写入文件
		if err := os.WriteFile(userConfigPath, data, 0600); err != nil {
			return fmt.Errorf("error writing config for user %s: %w", user, err)
		}
	}
	return nil
}

// hosts format is {user}@{ip}
// left usePasswd to "" if you want to use key
// return output if success
func StartRemoteCmds(hosts []string, remoteCmd string, usePasswd string) []string {
	fmt.Println()
	Logger.Debugf("Starting remote command: %s", remoteCmd)

	// 如果提供了密码，为每个用户创建配置文件
	if usePasswd != "" {
		if err := createUserConfigs(hosts, usePasswd); err != nil {
			Logger.Warnf("Failed to create user configs: %v", err)
		}
	}

	runRemoteCommand := func(host string, index int, logFile string, ch chan<- NodeMsg, remote_cmd string) {
		// 打开日志文件（追加模式），确保在函数退出时关闭文件
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error opening log file: %v", err), Complete: true}
			return
		}
		defer file.Close()

		file.WriteString(fmt.Sprintf("Running command on host %s:\n  %s\n", host, remote_cmd))

		// 模拟运行远程指令，实际可以替换为 SSH 或其他工具
		hostsplit := strings.Split(host, "@")
		user := hostsplit[0]
		server := hostsplit[1]
		client, session, err := sshSession(server, user, usePasswd)
		if err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error ssh: %v", err), Complete: true}
			return
		}

		// 1. 准备远程目录
		// PrintStep("StartRemoteCmds", color.BlueString("prepare remote %s secret directory", host))
		ch <- NodeMsg{Index: index, Output: fmt.Sprintf("prepare remote %s secret directory", host), Complete: false}
		prepareDirCmd := fmt.Sprintf("mkdir -p /teledeploy_secret/config && chown %s:%s /teledeploy_secret/config && chmod 700 /teledeploy_secret/config", user, user)
		if err := session.Run(prepareDirCmd); err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error preparing directory: %v", err), Complete: true}
			return
		}

		// 2. 配置 rclone
		// PrintStep("StartRemoteCmds", color.BlueString("configure rclone for %s", host))
		ch <- NodeMsg{Index: index, Output: fmt.Sprintf("configure rclone for %s", host), Complete: false}
		rcloneName := base64.RawURLEncoding.EncodeToString([]byte(server))
		err = NewRcloneConfiger(RcloneConfigTypeSsh{}, rcloneName, server).
			WithUser(user, usePasswd).
			DoConfig()
		if err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error configuring rclone: %v", err), Complete: true}
			return
		}

		// 3. 传输配置文件
		// PrintStep("StartRemoteCmds", color.BlueString("transfer config file to %s", host))
		ch <- NodeMsg{Index: index, Output: fmt.Sprintf("transfer config file to %s", host), Complete: false}
		localConfigPath := GetCurUserConfigPath()
		remoteConfigPath := fmt.Sprintf("/teledeploy_secret/config/userconfig_%s", user)
		// 使用 rclone 传输文件
		ch <- NodeMsg{Index: index, Output: fmt.Sprintf("transfer config file to %s", host), Complete: false}
		if err := RcloneSyncFileToFile(localConfigPath, fmt.Sprintf("%s:%s", rcloneName, remoteConfigPath)); err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error transferring config: %v", err), Complete: true}
			return
		}

		// // 4. 设置配置文件权限
		// PrintStep("StartRemoteCmds", color.BlueString("set ssh config permissions for %s", host))
		chmodCmd := fmt.Sprintf("chmod 600 %s", remoteConfigPath)
		if err := session.Run(chmodCmd); err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error setting config permissions: %v", err), Complete: true}
			return
		}

		// 创建管道用于读取 stdout 和 stderr
		stdoutPipe, _ := session.StdoutPipe()
		stderrPipe, _ := session.StderrPipe()

		// 合并 stdout 和 stderr
		reader := io.MultiReader(stdoutPipe, stderrPipe)

		// 同时输出到日志文件和通道
		writer := io.MultiWriter(file)

		// 创建 Scanner 读取合并流
		scanner := bufio.NewScanner(reader)

		// 启动命令
		if err := session.Start(remote_cmd); err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error starting command: %v", err), Complete: true}
			return
		}

		// 扫描合并后的输出流
		for scanner.Scan() {
			line := scanner.Text()
			ch <- NodeMsg{Index: index, Output: line, Complete: false}

			// 将输出写入日志文件
			_, _ = writer.Write([]byte(line + "\n"))
		}

		// 等待命令完成
		if err := session.Wait(); err != nil {
			ch <- NodeMsg{Index: index, Output: fmt.Sprintf("Error %v with %s", err, logFile), Complete: true}
			return
		}

		session.Close()
		client.Close()

		// 发送完成消息
		ch <- NodeMsg{Index: index, Output: DonePrefix + logFile, Complete: true}
	}

	// hosts := []string{"127.0.0.1", "8.8.8.8", "1.1.1.1"}
	nodes := make([]NodeState, len(hosts))
	for i, host := range hosts {
		nodes[i] = NodeState{Host: host}
	}

	model := RemoteControlModel{Nodes: nodes}
	program := tea.NewProgram(model)

	// 并行执行远程指令
	msgCh := make(chan NodeMsg)
	var wg sync.WaitGroup
	t := timestamp()
	for i, host := range hosts {
		wg.Add(1)
		go func(index int, host string) {
			defer wg.Done()
			path0 := ""
			for {
				randtag := rand.Intn(10000)
				if _, err := os.Stat(path0); err != nil {
					path0 = path.Join(LogDir(), "remote_cmd_"+strings.ReplaceAll(host, "@", "_")+"_"+t+"_"+fmt.Sprintf("%v", randtag)+".log")
					break
				}
			}

			runRemoteCommand(host, index, path0, msgCh, remoteCmd)
		}(i, host)
	}

	go func() {
		wg.Wait()
		close(msgCh)
	}()

	outputs := make([]string, len(hosts))
	matchUserIdx := func(output string) int {
		for i, host := range hosts {
			hostFormated := strings.ReplaceAll(host, "@", "_")
			if strings.Contains(output, hostFormated) {
				return i
			}
		}
		return -1
	}
	// 启动消息监听
	go func() {
		for msg := range msgCh {
			if strings.Contains(msg.Output, DonePrefix) {
				idx := matchUserIdx(msg.Output)
				if idx >= 0 {
					outputs[idx] = strings.Replace(msg.Output, DonePrefix, "", 1)
					// read from path
					if outputs[idx] != "" {
						contentbytes, err := os.ReadFile(outputs[idx])
						content := string(contentbytes)
						lines := strings.Split(string(content), "\n")
						if len(lines) >= 2 {
							// remove first two line
							content = strings.Join(lines[2:], "\n")
						}
						if err != nil {
							Logger.Warnf("Error reading log file(%s): %v", outputs[idx], err)
						} else {
							outputs[idx] = content
						}
					} else {
						Logger.Warnf("Remote output is empty")
					}
				} else {
					Logger.Warnf("Remote run maybe failed, output: %s", msg.Output)
				}
			}
			program.Send(msg)
		}
	}()

	// 启动 TUI
	if err := program.Start(); err != nil {
		fmt.Println("Error starting program:", err)
	}

	return outputs
}
