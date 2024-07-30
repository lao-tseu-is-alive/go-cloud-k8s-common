package tools

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func ToSnakeCase(str string) string {
	var result = ""
	var previousIsNumber = false
	var previousIsCapital bool = false
	for i, c := range strings.TrimSpace(str) {
		if i > 0 && unicode.IsUpper(c) && !previousIsNumber {
			if !previousIsCapital {
				result += "_"
			}
		}
		if unicode.IsUpper(c) {
			previousIsCapital = true
		} else {
			previousIsCapital = false
		}
		if unicode.IsNumber(c) {
			previousIsNumber = true
		} else {
			previousIsNumber = false
		}
		result += string(c)
	}
	return strings.ToLower(result)
}

func ToKebabCase(str string) string {
	res := strings.TrimSpace(str)
	snake := matchFirstCap.ReplaceAllString(res, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}
