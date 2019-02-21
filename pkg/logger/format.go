package logger

import (
	"strings"
)

type FitTextOptions struct {
	ExtraIndentWidth int
	MaxWidth         int
	MarkWrappedLine  bool
}

func FitText(text string, options FitTextOptions) string {
	tw := TerminalWidth()
	var lineWidth int
	if options.MaxWidth != 0 && tw > options.MaxWidth {
		lineWidth = options.MaxWidth
	} else {
		lineWidth = tw
	}

	return fitTextWithIndent(text, lineWidth, options.ExtraIndentWidth, options.MarkWrappedLine)
}

func fitTextWithIndent(text string, lineWidth, extraIndentWidth int, markWrappedLine bool) string {
	var result string
	var resultLines []string

	contentWidth := lineWidth - terminalServiceWidth() - extraIndentWidth
	fittedText := fitText(text, 0, contentWidth, markWrappedLine)
	for _, line := range strings.Split(fittedText, "\n") {
		indent := strings.Repeat(" ", extraIndentWidth)
		resultLines = append(resultLines, strings.Join([]string{indent, line}, ""))
	}

	result = strings.Join(resultLines, "\n")

	return result
}

func fitText(text string, contentCursor int, contentWidth int, markWrappedLine bool) string {
	if markWrappedLine {
		contentWidth -= 2
	}

	var fittedText string
	var line, word string
	var wordCursor int

	lineCursor := contentCursor

	wrappedLine := func() string {
		if markWrappedLine {
			if contentWidth-lineCursor+1 > 0 {
				line += strings.Repeat(" ", contentWidth-lineCursor+1)
			} else {
				line += " "
			}

			line += "↵"
		}

		return line
	}

	for _, r := range []rune(text) {
		word += string(r)
		wordCursor := lineCursor + len([]rune(word))

		switch string(r) {
		case "\n", "\r":
			if wordCursor > contentWidth && line != "" {
				fittedText += wrappedLine()
				fittedText += "\n"
				fittedText += word
			} else {
				fittedText += line + word
			}

			line = ""
			lineCursor = 0
			word = ""
			wordCursor = 0
		case " ":
			if wordCursor > contentWidth && line != "" {
				fittedText += wrappedLine()
				fittedText += "\n"
				line = ""
				lineCursor = 0
			}

			line += word
			lineCursor += len([]rune(word))
			word = ""
			wordCursor = 0
		}
	}

	if line != "" || word != "" {
		wordCursor = lineCursor + len([]rune(word))
		if wordCursor > contentWidth {
			fittedText += wrappedLine()

			if word != "" {
				fittedText += "\n"
				fittedText += word
			}
		} else {
			fittedText += line
			fittedText += word
		}
	}

	return fittedText
}
