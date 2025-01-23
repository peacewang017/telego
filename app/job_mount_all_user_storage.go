package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"telego/util"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ModJobMountAllUserStorageStruct struct{}

var ModJobMountAllUserStorage ModJobMountAllUserStorageStruct

func (m ModJobMountAllUserStorageStruct) JobCmdName() string {
	return "mount-all-user-storage"
}

// 请求结构体
// 这里假设一个用户在 Gemini 等平台有相同的用户名与密码
type GetAllUserStorageLinkRequest struct {
	UserName string `json:"username"` // 第三方平台用户名
	PassWord string `json:"password"` // 第三方平台密码
}

// 返回结构体
type GetAllUserStorageLinkResponse struct {
	RemoteInfos []util.UserMountsInfo `json:"remote_infos"` // 用于挂载的远程链接信息
}

func (m ModJobMountAllUserStorageStruct) inputUserLoginInfo() (username, password string, err error) {
	var ok bool
	ok, username = util.StartTemporaryInputUI(
		color.GreenString("ModJobMountAllUserStorageStruct.Run: Mount 用户存储空间需要鉴权，userName"),
		"输入 userName",
		"回车确认，ctrl + c 取消",
	)
	if !ok {
		fmt.Println("User canceled input")
		err = fmt.Errorf("ModJobMountAllUserStorageStruct.inputUserLoginInfo: error input")
		return
	}

	ok, password = util.StartTemporaryInputUI(
		color.GreenString("ModJobMountAllUserStorageStruct.Run: Mount 用户存储空间需要鉴权，passWord"),
		"输入 passWord",
		"回车确认，ctrl + c 取消",
	)
	if !ok {
		fmt.Println("User canceled input")
		err = fmt.Errorf("ModJobMountAllUserStorageStruct.inputUserLoginInfo: error input")
		return
	}
	return
}

func (m ModJobMountAllUserStorageStruct) getUserLoginInfo() (string, string, error) {
	var username, password string
	// 登录验证
	userLoginInfoFile := path.Join(util.WorkspaceDir(), "config/userinfo")
	if _, err := os.Stat(userLoginInfoFile); err == nil {
		// // 文件存在，尝试读取文件中的内容
		var content []byte
		content, err = os.ReadFile(userLoginInfoFile)
		if err != nil {
			return username, password, fmt.Errorf("m.getUserLoginInfo: %v", err)
		}

		// // 解析文件内容，查找 username 和 password
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line) // 去掉行首尾的空格
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "username:") {
				username = strings.TrimSpace(strings.TrimPrefix(line, "username:"))
			} else if strings.HasPrefix(line, "password:") {
				password = strings.TrimSpace(strings.TrimPrefix(line, "password:"))
			}
		}

		// // 如果从文件中读取了有效的 username 和 password，则直接使用
		if username != "" && password != "" {
			return username, password, nil
		} else {
			fmt.Println("ModJobMountAllUserStorageStruct.Run: No valid credentials found in userinfo file, using UI for input.")
			username, password, err = m.inputUserLoginInfo()
			if err != nil {
				return username, password, fmt.Errorf("ModJobMountAllUserStorageStruct.getUserLoginInfo: %v", err)
			}
		}
	} else {
		fmt.Println("ModJobMountAllUserStorageStruct.Run: No valid credentials found in userinfo file, using UI for input.")
		username, password, err = m.inputUserLoginInfo()
		if err != nil {
			return username, password, fmt.Errorf("ModJobMountAllUserStorageStruct.getUserLoginInfo: %v", err)
		}
	}
	return username, password, nil
}

func (m ModJobMountAllUserStorageStruct) saveUserLoginInfo(username, password string) error {
	userLoginInfoFile := path.Join(util.WorkspaceDir(), "config/userinfo")
	err := os.WriteFile(userLoginInfoFile, []byte("username: "+username+"\npassword: "+password+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("ModJobMountAllUserStorageStruct.saveUserLoginInfo: Error writing to userinfo file: %v", err)
	}
	return nil
}

func (m ModJobMountAllUserStorageStruct) getLocalRootStorage() (string, error) {
	// 选定本地挂载根目录
	ok, localRootStorage := util.StartTemporaryInputUI(
		color.GreenString("选定本地挂载根目录"),
		"输入可用本地路径",
		"回车确认，ctrl + c 取消",
	)
	if !ok {
		fmt.Println("User canceled input")
		return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: Error user keyboard input")
	}

	// 对本地挂载根目录进行检验
	if _, err := os.Stat(localRootStorage); os.IsNotExist(err) {
		err := os.Mkdir(localRootStorage, 0755)
		if err != nil {
			return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: Error creating %s", localRootStorage)
		}
	} else {
		// // 检查是否是文件
		info, err := os.Stat(localRootStorage)
		if err != nil {
			return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: Error opening %s", localRootStorage)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: %s is a file, not a directory", localRootStorage)
		}

		files, err := os.ReadDir(localRootStorage)
		if err != nil {
			return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: Error opening %s", localRootStorage)
		}

		if len(files) != 0 {
			return "", fmt.Errorf("ModJobMountAllUserStorageStruct.getLocalRootStorage: %s not empty", localRootStorage)
		}
	}

	return localRootStorage, nil
}

func (m ModJobMountAllUserStorageStruct) Run() {
	username, password, err := m.getUserLoginInfo()
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: %v", err)
	}

	// 初始化 http 请求
	req := GetAllUserStorageLinkRequest{
		UserName: username,
		PassWord: password,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error marshalling request: %v", err)
	}

	// 执行请求，拿到 AllUserStorageLink
	serverUrl, err := (util.MainNodeConfReader{}).ReadPubConf(util.PubConfMountAllUserStorageServerUrl{})
	serverUrl = strings.TrimSpace(serverUrl)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Get server ip error")
		return
	}
	client := &http.Client{Timeout: 60 * time.Second} // 设置 60 秒超时
	httpResp, err := client.Post(util.UrlJoin(serverUrl, "/mount_all_user_storage_server_url"), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error getting response (server_url: %s, fullUrl: %s)", serverUrl, util.UrlJoin(serverUrl, "/mount_all_user_storage_server_url"))
		return
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body) // 读取错误信息
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Unexpected status: %d, %s", httpResp.StatusCode, string(body))
		return
	}
	var resp GetAllUserStorageLinkResponse
	decoder := json.NewDecoder(httpResp.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error decoding response: %v", err)
		return
	}

	// 用户指定本地挂载点
	localRootStorage, err := m.getLocalRootStorage()
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error asserting local mount point: %v", err)
	}

	// 调用 SshFs / rclone 进行挂载
	// // 检查安装 SshFs 和 Rclone
	if runtime.GOOS == "linux" && !(BinManagerSshFs{}).CheckInstalled() {
		err := NewBinManager(BinManagerSshFs{}).MakeSureWith()
		if err != nil {
			fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error making sure sshfs: %v", err)
		}
	} else if runtime.GOOS == "windows" && (BinManagerRclone{}).CheckInstalled() {
		err := NewBinManager(BinManagerRclone{}).MakeSureWith()
		if err != nil {
			fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error making sure rclone: %v", err)
		}
	}

	// // 分系统挂载
	// for _, userMountInfo := range resp.RemoteInfos {
	// 	if runtime.GOOS == "linux" {
	// 		// （未确定）
	// 		sshFsArgv := &sshFsArgv{
	// 			remotePath: userMountInfo.AccessServer + ":" + userMountInfo.UserStorage_.RootStorage,
	// 			localPath:  path.Join(localRootStorage, userMountInfo.UserStorage_.Name()),
	// 		}
	// 		ModJobSshFs.doMount(sshFsArgv)

	// 	} else if runtime.GOOS == "windows" {
	// 		rCloneConfiger := util.NewRcloneConfiger(util.RcloneConfigTypeSsh{}, userMountInfo.UserStorage_.Name(), userMountInfo.AccessServer)
	// 		rCloneConfiger.WithUser(userName, passWord).DoConfig()

	// 		// (未实现)
	// 	}
	// }
	fmt.Print("\n--------------------------------------------------\n")
	fmt.Printf("Local mountpoint: %s", localRootStorage)
	for idx, userMountInfo := range resp.RemoteInfos {
		fmt.Printf("Storage-%d\n", idx)
		fmt.Printf("type: %s, access-server: %s, root-path: %s\n", userMountInfo.UserStorage_.Type, userMountInfo.AccessServer, userMountInfo.UserStorage_.RootStorage)
		for idx2, subPath := range userMountInfo.UserStorage_.SubPaths {
			fmt.Printf("subpath-%d: %s\n", idx2, subPath)
		}
	}
	fmt.Print("\n--------------------------------------------------\n")

	// 写入配置文件
	err = m.saveUserLoginInfo(password, username)
	if err != nil {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Error saving user login info: %v", err)
	} else {
		fmt.Printf("ModJobMountAllUserStorageStruct.Run: Configuration saved successfully.")
	}
}

func (m ModJobMountAllUserStorageStruct) ParseJob(mountAllUserStorageCmd *cobra.Command) *cobra.Command {
	mountAllUserStorageCmd.Run = func(_ *cobra.Command, _ []string) {
		m.Run()
	}
	return mountAllUserStorageCmd
}
