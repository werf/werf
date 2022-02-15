package util

import (
	"bufio"
	"strings"
)

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
