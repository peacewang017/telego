package util

import "strings"

type UserOneStorageSet struct {
	// Type 区分不同的集群
	// RootStorage 区分集群内不同的存储
	// SubPaths 区分存储内不同的用户目录
	Type        string   // "gemini"
	RootStorage string   // "/gemini-sh"
	SubPaths    []string // ["user/lzy", "share/fasdfs"]
}

func (m UserOneStorageSet) Name() string {
	// trim header and tailing /
	return strings.Trim(m.RootStorage, "/")
}
