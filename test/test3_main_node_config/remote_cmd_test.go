package test3_main_node_config

import (
	"os"
	"telego/util"
	"testing"
)

func TestRemoteCmd(t *testing.T) {
	res, logfile := util.StartRemoteCmds([]string{
		util.MainNodeUser + "@" + util.MainNodeIp + ":" + util.MainNodeSshPort}, "echo helloworld", util.MainNodeUser)
	// if lines < 3, fatal with log
	{
		lines := len(res[0])
		if lines < 3 {
			// 读取日志文件内容，使用标准库
			logContent, readErr := os.ReadFile(logfile[0])
			if readErr != nil {
				t.Logf("读取日志文件失败: %v", readErr)
			}
			t.Fatalf("远程命令执行失败，返回行数不足，预期至少3行，实际%d行。日志内容：%s", lines, string(logContent))
		}

		// 验证命令输出包含预期结果
		foundHelloWorld := false
		for _, lineRune := range res[0] {
			// 将rune转换为string进行比较
			line := string(lineRune)
			if line == "helloworld" {
				foundHelloWorld = true
				break
			}
		}

		if !foundHelloWorld {
			// 读取日志文件内容，使用标准库
			logContent, readErr := os.ReadFile(logfile[0])
			if readErr != nil {
				t.Logf("读取日志文件失败: %v", readErr)
			}
			// 打印日志内容而不只是文件名
			t.Fatalf("远程命令未返回预期输出 'helloworld'。完整输出：%v，日志内容：%s", res[0], string(logContent))
		}
	}
}
