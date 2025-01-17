package util

import "strings"

type UserOneStorageSet struct {
	// 对于一个sftp来说，路径名不可能重名，所以可以用作uniqueid
	// 根据 main_node 上存储名称来，例如 gemini-nm / gemini-sh
	RootStorage string
	SubPaths    []string
}

func (m UserOneStorageSet) Name() string {
	// trim header and tailing /
	return strings.Trim(m.RootStorage, "/")
}
