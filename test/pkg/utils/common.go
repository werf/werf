package utils

import (
	"bufio"
	"strings"

	. "github.com/onsi/gomega"
)

func StringToLines(s string) (lines []string) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	Expect(scanner.Err()).Should(Succeed())

	return
}
