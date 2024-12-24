package util

import "strings"

func PathWinStyleToLinux(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

func PathIsAbsolute(path string) bool {
	return path[0] == '/' || strings.Contains(path, ":/") || strings.Contains(path, ":\\")
}

func PathIsRelative(path string) bool {
	return !PathIsAbsolute(path)
}
