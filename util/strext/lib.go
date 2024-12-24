package strext

func SafeSubstring(str string, start, length int) string {
	// 计算从 start 开始的字符的位置
	runes := []rune(str) // 将字符串转换为[]rune，这样可以按字符访问
	if start >= len(runes) {
		return ""
	}
	// 确保长度不超过可用字符长度
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
