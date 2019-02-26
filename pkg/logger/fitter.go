package logger

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

	fittedText, _ := fitText(text, fitterState{}, contentWidth, markWrappedLine, false)
	for _, line := range strings.Split(fittedText, "\n") {
		indent := strings.Repeat(" ", extraIndentWidth)
		resultLines = append(resultLines, strings.Join([]string{indent, line}, ""))
	}

	result = strings.Join(resultLines, "\n")

	return result
}

const (
	isControlSequenceNoneProcessed = iota
	isControlSequenceEscapeSequenceProcessed
	isControlSequenceOpenBorderProcessed
	isControlSequenceParametersProcessed

	resetColorControlSequence = "\x1b[0m"
	escapeSequenceCode        = 27

	resetColorCode = 0
)

type fitterState struct {
	wrapperState
	controlSequenceState
	colorState
}

type wrapperState struct {
	line       string
	word       string
	lineTWidth int
	wordTWidth int
}

type controlSequenceState struct {
	controlSequenceBytes       []rune
	controlSequenceCursorState int
}

type colorState struct {
	isColorLine               bool
	prevCursorRune            rune
	colorControlSequenceCodes []int
}

func (s *fitterState) lineAppend(value string, tWidth int) {
	s.line += value
	s.lineTWidth += tWidth
}

func (s *fitterState) wordAppend(value string, tWidth int) {
	s.word += value
	s.wordTWidth += tWidth
}

func (s *fitterState) lineWithWord() string {
	return s.line + s.word
}

func (s *fitterState) lineWithWordTWidth() int {
	return s.lineTWidth + s.wordTWidth
}

func (s *fitterState) lineReset() {
	s.line = ""
	s.lineTWidth = 0
}

func (s *fitterState) wordReset() {
	s.word = ""
	s.wordTWidth = 0
}

func (s *fitterState) filledLine(contentWidth int, markWrappedLine bool) string {
	if markWrappedLine {
		var padding int
		if s.lineTWidth <= contentWidthWithoutMarkSign(contentWidth, markWrappedLine) {
			padding = contentWidth - s.lineTWidth - 1
		}

		return s.line + strings.Repeat(" ", padding) + "↵"
	} else {
		return s.line
	}
}

func (s *fitterState) parseColorCodes() []int {
	preparedColorFormatsPart := string(s.controlSequenceBytes[:len(s.controlSequenceBytes)-1])
	preparedColorFormatsPart = string(bytes.TrimPrefix([]byte(preparedColorFormatsPart), []byte{escapeSequenceCode, []byte("[")[0]}))

	colorCodesStrings := strings.Split(preparedColorFormatsPart, ";")
	var colorCodes []int
	for _, colorCodeString := range colorCodesStrings {
		if colorCodeString == "" {
			continue
		}

		cd, err := strconv.Atoi(colorCodeString)
		if err != nil {
			panic(err)
		}

		colorCodes = append(colorCodes, cd)
	}

	return colorCodes
}

func (s *fitterState) generateColorControlSequence() string {
	var result string

	if len(s.colorControlSequenceCodes) != 0 {
		result = "\x1b["

		var colorCodesStrings []string
		for _, colorCode := range s.colorControlSequenceCodes {
			colorCodesStrings = append(colorCodesStrings, fmt.Sprintf("%d", colorCode))
		}
		result += strings.Join(colorCodesStrings, ";")

		result += "m"
	}

	return result
}

func (s *fitterState) resetColorCodes() {
	s.colorControlSequenceCodes = []int{}
}

func contentWidthWithoutMarkSign(contentWidth int, markWrappedLine bool) int {
	if markWrappedLine {
		return contentWidth - 1
	}

	return contentWidth
}

func fitText(text string, s fitterState, contentWidth int, markWrappedLine bool, cacheIncompleteLine bool) (string, fitterState) {
	var result string

	for _, r := range []rune(text) {
		result += runFitterWrapper(r, &s, contentWidth, markWrappedLine)
		ignoreControlSequenceTWidth(r, &s)
	}

	if !cacheIncompleteLine {
		result += processFitterCachedLineAndWord(&s, contentWidth, markWrappedLine)
	}

	result = addRequiredColorControlSequences(result, &s)

	return result, s
}

func runFitterWrapper(r rune, s *fitterState, contentWidth int, markWrappedLine bool) string {
	var result string

	switch string(r) {
	case "\b":
		if s.wordTWidth != 0 {
			s.wordTWidth -= 1
			s.wordAppend("\b", 0)
		} else if s.lineTWidth != 0 {
			s.lineTWidth -= 1
			s.lineAppend("\b", 0)
		}
	case "\n", "\r":
		carriage := string(r)

		if s.lineWithWordTWidth() <= contentWidth {
			result += s.lineWithWord()
		} else if s.lineWithWordTWidth() > contentWidth {
			result += s.filledLine(contentWidth, markWrappedLine)
			result += "\n"
			result += s.word
		}

		result += carriage

		s.lineReset()
		s.wordReset()
	case " ":
		space := string(r)
		spaceTWidth := len(space)

		if s.lineWithWordTWidth()+spaceTWidth > contentWidthWithoutMarkSign(contentWidth, markWrappedLine) {
			result += s.filledLine(contentWidth, markWrappedLine)
			result += "\n"

			s.lineReset()
		}

		s.lineAppend(s.word+space, s.wordTWidth+spaceTWidth)
		s.wordReset()
	default:
		s.wordAppend(string(r), 1)
	}

	return result
}

func ignoreControlSequenceTWidth(r rune, s *fitterState) {
	processFitterControlSequence(r, s, nil)
}

func processFitterControlSequence(r rune, s *fitterState, processColorControlSequenceFunc func(f *fitterState)) {
	switch s.controlSequenceCursorState {
	case isControlSequenceNoneProcessed:
		switch r {
		case escapeSequenceCode:
			s.controlSequenceBytes = []rune{r}
			s.controlSequenceCursorState = isControlSequenceEscapeSequenceProcessed
		}
	case isControlSequenceEscapeSequenceProcessed:
		switch string(r) {
		case "[":
			s.controlSequenceBytes = append(s.controlSequenceBytes, r)
			s.controlSequenceCursorState = isControlSequenceOpenBorderProcessed
		}
	case isControlSequenceOpenBorderProcessed, isControlSequenceParametersProcessed:
		if unicode.IsNumber(r) || string(r) == ";" {
			s.controlSequenceBytes = append(s.controlSequenceBytes, r)
			s.controlSequenceCursorState = isControlSequenceParametersProcessed
		} else {
			if unicode.IsLetter(r) {
				s.controlSequenceBytes = append(s.controlSequenceBytes, r)
				if s.wordTWidth-len(s.controlSequenceBytes) >= 0 {
					s.wordTWidth -= len(s.controlSequenceBytes)
				} else {
					s.wordTWidth = 0
				}

				if string(r) == "m" && processColorControlSequenceFunc != nil {
					processColorControlSequenceFunc(s)
				}
			}

			s.controlSequenceCursorState = isControlSequenceNoneProcessed
		}
	default:
		s.controlSequenceCursorState = isControlSequenceNoneProcessed
	}
}

func processFitterCachedLineAndWord(s *fitterState, contentWidth int, markWrappedLine bool) string {
	var result string

	if s.lineWithWordTWidth() > contentWidthWithoutMarkSign(contentWidth, markWrappedLine) {
		if s.line != "" {
			result += s.filledLine(contentWidth, markWrappedLine)

			if s.word != "" {
				result += "\n"
				result += s.word

				s.lineReset()
				s.wordReset()
			}
		} else if s.word != "" {
			result += s.word

			s.wordReset()
		}
	} else {
		result += s.lineWithWord()
	}

	return result
}

func addRequiredColorControlSequences(fittedText string, s *fitterState) string {
	var result string

	for _, r := range []rune(fittedText) {
		switch string(r) {
		case "\n", "\r":
			if string(s.prevCursorRune) == "\r" {
				result += string(r)
			} else {
				if s.isColorLine {
					result += resetColorControlSequence
				}

				result += string(r)
			}
		default:
			if string(s.prevCursorRune) == "\r" || string(s.prevCursorRune) == "\n" {
				result += s.generateColorControlSequence()
			}

			result += string(r)
		}

		s.prevCursorRune = r
		processFitterControlSequence(r, s, processColorControlSequence)
	}

	return result
}

func processColorControlSequence(s *fitterState) {
	colorCodes := s.parseColorCodes()
	for _, colorCode := range colorCodes {
		if isResetColorCode(colorCode) {
			s.resetColorCodes()
			s.isColorLine = false
		} else {
			s.colorControlSequenceCodes = append(s.colorControlSequenceCodes, colorCode)
			s.isColorLine = true
		}
	}
}

func isResetColorCode(code int) bool {
	return code == resetColorCode
}
