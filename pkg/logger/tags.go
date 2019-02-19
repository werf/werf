package logger

import (
	"fmt"
	"strings"
)

var (
	tagValue        = ""
	tagColorizeFunc = func(a ...interface{}) string { return fmt.Sprint(a...) }
	tagIndentWidth  = 2
)

func WithTag(value string, colorizeFunc func(...interface{}) string, f func() error) error {
	savedTag := tagValue
	savedColorizeFunc := tagColorizeFunc
	SetTag(value, colorizeFunc)
	err := f()
	SetTag(savedTag, savedColorizeFunc)
	return err
}

func SetTag(value string, colorizeFunc func(...interface{}) string) {
	tagValue = value
	tagColorizeFunc = colorizeFunc
}

func formattedTag() string {
	if len(tagValue) == 0 {
		return ""
	}

	return strings.Join([]string{
		tagColorizeFunc(tagValue),
		strings.Repeat(" ", tagIndentWidth),
	}, "")
}

func tagBlockWidth() int {
	if len(tagValue) == 0 {
		return 0
	}

	return len(tagValue) + tagIndentWidth
}
