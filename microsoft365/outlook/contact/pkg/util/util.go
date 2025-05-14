package util

import "strings"

func Deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

func SplitString(s string) []string {
	res := strings.Split(s, ",")
	if len(res) == 1 && res[0] == "" {
		return []string{}
	}
	return res
}
