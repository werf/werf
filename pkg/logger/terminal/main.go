package terminal

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	defaultTerminalWidth = 120
)

func Width() int {
	if wtw, ok := os.LookupEnv("WERF_TERMINAL_WIDTH"); ok {
		if i, err := strconv.Atoi(wtw); err != nil {
			panic(fmt.Sprintf("Unexpected WERF_TERMINAL_WIDTH: %s", err))
		} else {
			return i
		}
	} else {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				panic(err)
			}

			return w
		}
	}

	return defaultTerminalWidth
}

func FitTextWithIndent(text string, indentWidth int) string {
	return fitTextWithIndent(text, Width(), indentWidth)
}

func FitTextWithIndentWithWidthMaxLimit(text string, indentWidth int, maxWidth int) string {
	tw := Width()
	var lineWidth int
	if tw < maxWidth {
		lineWidth = tw
	} else {
		lineWidth = maxWidth
	}

	return fitTextWithIndent(text, lineWidth, indentWidth)
}

func fitTextWithIndent(text string, lineWidth, indentWidth int) string {
	var result string
	var resultLines []string

	contentWidth := lineWidth - indentWidth
	fittedText := fitText(text, contentWidth)
	for _, line := range strings.Split(fittedText, "\n") {
		indent := strings.Repeat(" ", indentWidth)
		resultLines = append(resultLines, strings.Join([]string{indent, line}, ""))
	}

	result = strings.Join(resultLines, "\n")

	return result
}

func fitText(text string, contentWidth int) string {
	var result string
	var resultLines []string

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		var cursor int
		var resultLine string

		lineWords := strings.Split(line, " ")
		if len(lineWords) == 1 && len(lineWords[0]) == 0 {
			resultLines = append(resultLines, "")
		} else {
			for ind, word := range lineWords {
				isLastWord := ind == len(lineWords)-1

				toAdd := word
				if !isLastWord {
					toAdd += " "
				}

				if cursor+len(toAdd) > contentWidth && resultLine != "" {
					resultLines = append(resultLines, resultLine)
					cursor = 0
					resultLine = ""
				}

				cursor += len(toAdd)
				resultLine += toAdd
			}

			if resultLine != "" {
				resultLines = append(resultLines, resultLine)
			}
		}
	}

	result = strings.Join(resultLines, "\n")

	return result
}
