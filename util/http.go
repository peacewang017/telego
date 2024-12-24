package util

import (
	"fmt"
	"io"
	"net/http"
)

func HttpGetUrlContent(url string) ([]byte, error) {
	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		// fmt.Printf("Error fetching file: %v\n", err)
		Logger.Warnf("Failed to fetch file: %v", url)
		return []byte{}, err
	}
	defer resp.Body.Close() // 确保响应体被正确关闭

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		// fmt.Printf("Failed to fetch file: %s\n", resp.Status)
		Logger.Warnf("Failed to fetch file: %s", resp.Status)
		return []byte{}, fmt.Errorf("failed to fetch file with StatusCode %s", resp.Status)
	}

	// 读取文件内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Printf("Error reading file: %v\n", err)
		Logger.Warnf("Failed to read http resp body, err: %v", err)
		return []byte{}, err
	}
	return data, nil
}
