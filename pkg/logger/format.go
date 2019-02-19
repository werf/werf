package logger

import (
	"strings"
)

type FitTextOptions struct {
	ExtraIndentWidth int
	MaxWidth         int
	MarkEndOfLine    bool
}

func FitText(text string, options FitTextOptions) string {
	tw := TerminalWidth()
	var lineWidth int
	if options.MaxWidth != 0 && tw > options.MaxWidth {
		lineWidth = options.MaxWidth
	} else {
		lineWidth = tw
	}

	return fitTextWithIndent(text, lineWidth, options.ExtraIndentWidth, options.MarkEndOfLine)
}

func fitTextWithIndent(text string, lineWidth, extraIndentWidth int, markEndOfLine bool) string {
	var result string
	var resultLines []string

	contentWidth := lineWidth - terminalServiceWidth() - extraIndentWidth
	fittedText := fitText(text, 0, contentWidth, markEndOfLine)
	for _, line := range strings.Split(fittedText, "\n") {
		indent := strings.Repeat(" ", extraIndentWidth)
		resultLines = append(resultLines, strings.Join([]string{indent, line}, ""))
	}

	result = strings.Join(resultLines, "\n")

	return result
}

func fitText(text string, cursor int, contentWidth int, markEndOfLine bool) string {
	if markEndOfLine {
		contentWidth -= 2
	}

	var fittedText string
	var line, word string
	var wordCursor int

	lineCursor := cursor

	addWithSignInLine := func() {
		if markEndOfLine {
			if contentWidth-lineCursor+1 > 0 {
				line += strings.Repeat(" ", contentWidth-lineCursor+1)
			} else {
				line += " "
			}

			line += "↵"
		}
	}

	for _, r := range []rune(text) {
		word += string(r)
		wordCursor := lineCursor + len([]rune(word))

		switch string(r) {
		case "\n", "\r":
			if wordCursor > contentWidth && line != "" {
				addWithSignInLine()

				fittedText += line
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
				addWithSignInLine()

				fittedText += line
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
			addWithSignInLine()

			fittedText += line

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
