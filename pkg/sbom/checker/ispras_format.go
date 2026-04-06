package checker

import "fmt"

type IsprasFormat string

const (
	IsprasFormatOSS       IsprasFormat = "oss"
	IsprasFormatContainer IsprasFormat = "container"
)

func (t IsprasFormat) String() string {
	return string(t)
}

func ParseIsprasFormat(s string) (IsprasFormat, error) {
	switch IsprasFormat(s) {
	case IsprasFormatOSS:
		return IsprasFormatOSS, nil
	case IsprasFormatContainer:
		return IsprasFormatContainer, nil
	default:
		return "", fmt.Errorf("invalid ispras format %q: must be one of %s, %s", s, IsprasFormatOSS, IsprasFormatContainer)
	}
}
