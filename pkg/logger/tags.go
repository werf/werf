package logger

import "strings"

var (
	colorlessTag   = ""
	tagWidth       = 20
	tagIndentWidth = 2
)

func WithTag(value string, f func() error) error {
	savedTag := colorlessTag
	colorlessTag = value
	err := f()
	colorlessTag = savedTag

	return err
}

func SetTag(value string) {
	colorlessTag = value
}

func formattedTag() string {
	var fittedTag string

	if len(colorlessTag) == 0 {
		return ""
	} else if len(colorlessTag) > tagWidth {
		longTagPostfix := " ..."
		fittedTag = colorlessTag[:tagWidth-len(longTagPostfix)] + longTagPostfix
	} else {
		fittedTag = colorlessTag
	}

	padLeft := strings.Repeat(" ", tagWidth-len(fittedTag))
	colorizedTag := colorize(fittedTag, tagFormat...)

	return strings.Join([]string{padLeft, colorizedTag, strings.Repeat(" ", tagIndentWidth)}, "")
}

func tagBlockWidth() int {
	if len(colorlessTag) == 0 {
		return 0
	} else {
		return tagWidth + tagIndentWidth
	}
}
