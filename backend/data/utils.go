package data

import "regexp"

// @Author spark
// @Date 2025/2/13 13:08
// @Desc
//-----------------------------------------------------------------------------------

// RemoveAllBlankChar  使用正则表达式移除字符串中的空白字符
func RemoveAllBlankChar(s string) string {
	return removeAllSpaces(s)
}
func removeAllSpaces(s string) string {
	re := regexp.MustCompile(`\s`)
	return re.ReplaceAllString(s, "")
}
