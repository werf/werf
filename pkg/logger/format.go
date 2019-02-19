package logger

import (
	"strings"
)

type FitTextOptions struct {
	ExtraIndentWidth int
	MaxWidth         int
}

func FitText(text string, options FitTextOptions) string {
	tw := TerminalWidth()
	var lineWidth int
	if options.MaxWidth != 0 && tw > options.MaxWidth {
		lineWidth = options.MaxWidth
	} else {
		lineWidth = tw
	}

	return fitTextWithIndent(text, lineWidth, options.ExtraIndentWidth)
}

func fitTextWithIndent(text string, lineWidth, extraIndentWidth int) string {
	var result string
	var resultLines []string

	contentWidth := lineWidth - terminalServiceWidth() - extraIndentWidth
	fittedText := fitText(text, contentWidth)
	for _, line := range strings.Split(fittedText, "\n") {
		indent := strings.Repeat(" ", extraIndentWidth)
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

				if cursor+len(toAdd) >= contentWidth && resultLine != "" {
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
