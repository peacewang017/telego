package app

import (
	"fmt"
	"os"
	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ModJobMountAllUserStorageStruct struct{}

var ModJobMountAllUserStorage ModJobMountAllUserStorageStruct

func (m ModJobMountAllUserStorageStruct) JobCmdName() string {
	return "mount-all-user-storage"
}

type PathRequest struct {
	UserName string
	PassWord string
}

type PathResponse struct {
	RemotePaths []string `json: "remote_paths"`
}

func (m ModJobMountAllUserStorageStruct) Run() {
	ok, userName := util.StartTemporaryInputUI(
		color.GreenString("Mount 用户存储空间需要鉴权，userName"),
		"输入 userName",
		"回车确认，ctrl + c 取消",
	)
	if !ok {
		fmt.Println("User caceled input")
		os.Exit(1)
	}

	ok, passWord := util.StartTemporaryInputUI(
		color.GreenString("Mount 用户存储空间需要鉴权，passWord"),
		"输入 passWord",
		"回车确认，ctrl + c 取消",
	)
	if !ok {
		fmt.Println("User caceled input")
		os.Exit(1)
	}

	// 进行 http 请求 server

	// 接收结果，调用 ssh / rclone 进行挂载
}

func (m ModJobMountAllUserStorageStruct) ParseJob(mountAllUserStorageCmd *cobra.Command) *cobra.Command {
	mountAllUserStorageCmd.Run = func(_ *cobra.Command, _ []string) {
		m.Run()
	}

	return mountAllUserStorageCmd
}
