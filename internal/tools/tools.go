package tools

import "strings"

// converts `error` to `string` while trimming any leading or trailing whitespace
func ErrStr(err error) string {
	return strings.TrimSpace(err.Error())
}
