package util

import (
	"fmt"
	"strings"
)

func NumerateLines(text string, firstLineNumber int) string {
	var res string

	lines := strings.Split(text, "\n")
	for lineInd, line := range lines {
		res += fmt.Sprintf("%6d  %s\n", firstLineNumber+lineInd, line)
	}

	return res
}
