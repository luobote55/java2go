package match

import (
	"regexp"
)

func Remove(str, remove string) string {
	return str
	//	return strings.Replace(str, remove, "", -1)
}

// 反引号匹配
func FindBacktick(str string) string {
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllString(str, -1)
	return matches[0]
}

// 反引号匹配
func FindBackticks(str string) []string {
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllString(str, -1)
	return matches
}

// 单引号匹配
func FindQuote(str string) string {
	re := regexp.MustCompile("('(.*?)')")
	matches := re.FindAllString(str, -1)
	return matches[0]
}

// 双引号匹配
func FindDoubleQuote(str string) string {
	re := regexp.MustCompile("(\"(.*?)\")")
	matches := re.FindAllString(str, -1)
	return matches[0]
}

// 双引号匹配
func FindDoubleQuotes(str string) []string {
	re := regexp.MustCompile("(\"(.*?)\")")
	matches := re.FindAllString(str, -1)
	return matches
}

// 字符串匹配
func FindFix(str string, fix string) string {
	re := regexp.MustCompile(fix)
	matches := re.FindStringSubmatch(str)
	if len(matches) == 0 {
		return ""
	}
	if len(matches) == 1 {
		return ""
	}
	return matches[1]
}

// 字符串匹配
func FindFixs(str string, fix string) []string {
	re := regexp.MustCompile(fix)
	matches := re.FindStringSubmatch(str)
	return matches
}
