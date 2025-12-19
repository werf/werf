package dockerfile

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

// MapToCorrectHeredocCmd processes heredoc files by embedding their content into the command line.
// For files with Chomp flag, trailing newline characters (\r\n) are removed.
// Returns the updated full command line with heredoc content and a prepend shell flag
func MapToCorrectHeredocCmd(cmdLine instructions.ShellDependantCmdLine) (string, bool) {
	full := cmdLine.CmdLine[0]
	for _, file := range cmdLine.Files {
		name := file.Name
		data := file.Data
		isChomp := file.Chomp
		if isChomp {
			data = strings.TrimRight(data, "\r\n")
		}

		full += "\n" + data + name
	}

	return full, true
}

func InstructionUsesUnsupportedHeredoc(cmd InstructionDataInterface) bool {
	heredocPattern := `%s <<(-?)\s*([^<]*)$`

	var name, val string

	switch instrData := cmd.(type) {
	case *instructions.AddCommand:
		name = instrData.Name()
		val = instrData.String()
	case *instructions.CopyCommand:
		name = instrData.Name()
		val = instrData.String()
	default:
		return false
	}

	reHeredoc := regexp.MustCompile(fmt.Sprintf(heredocPattern, name))
	return len(reHeredoc.FindStringSubmatch(val)) > 0
}
