package data

import (
	"regexp"
	"strings"
)

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

// RemoveAllNonDigitChar 去除所有非数字字符
func RemoveAllNonDigitChar(s string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(s, "")
}

// RemoveAllDigitChar 去除所有数字字符
func RemoveAllDigitChar(s string) string {
	re := regexp.MustCompile(`\d`)
	return re.ReplaceAllString(s, "")
}

// ConvertStockCodeToTushareCode 将股票代码转换为tushare的股票代码
func ConvertStockCodeToTushareCode(stockCode string) string {
	//提取非数字
	stockCode = RemoveAllNonDigitChar(stockCode) + "." + strings.ToUpper(RemoveAllDigitChar(stockCode))
	return stockCode
}

// ConvertTushareCodeToStockCode 将tushare股票代码转换为的普通股票代码
func ConvertTushareCodeToStockCode(stockCode string) string {
	//提取非数字
	stockCode = strings.ToLower(RemoveAllDigitChar(stockCode)) + RemoveAllNonDigitChar(stockCode)
	return strings.ReplaceAll(stockCode, ".", "")
}
