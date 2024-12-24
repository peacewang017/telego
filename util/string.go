package util

import (
	"strings"
)

func Unquote(s string) string {

	ReplaceList := [][]string{
		{"\\n", "\n"},
		{"\\t", "\t"},
		{"\\b", "\b"},
		{"\\f", "\f"},
		{"\\r", "\r"},
		{"\\v", "\v"},
		{"\\a", "\a"},
		{"\\e", "\x1b"},
		{"\\0", "\x00"},
		{"\\u", "\u0000"},
		{"\\x", "\x00"},
		{"\\\\", "\\"},
		{"\\'", "'"},
		{"\\\"", "\""},
		{"\\377", "\xff"},
	}
	for _, v := range ReplaceList {
		s = strings.ReplaceAll(s, v[0], v[1])
	}
	return s
}
