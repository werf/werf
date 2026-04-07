package convert

import "fmt"

func NewAssembler(format string) (Assembler, error) {
	switch format {
	case "container":
		return &ISPRASContainerAssembler{}, nil
	case "oss":
		return &ISPRASOSSAssembler{}, nil
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}
