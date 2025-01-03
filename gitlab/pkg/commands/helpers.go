package commands

import "strings"

func IndentString(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}
