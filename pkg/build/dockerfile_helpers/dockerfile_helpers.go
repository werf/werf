package dockerfile_helpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

func ResolveDockerStagesFromValue(stages []instructions.Stage) {
	nameToIndex := make(map[string]string)
	for i, s := range stages {
		name := strings.ToLower(s.Name)
		index := strconv.Itoa(i)
		if name != index {
			nameToIndex[name] = index
		}

		for _, cmd := range s.Commands {
			switch typedCmd := cmd.(type) {
			case *instructions.CopyCommand:
				if typedCmd.From != "" {
					from := strings.ToLower(typedCmd.From)
					if val, ok := nameToIndex[from]; ok {
						typedCmd.From = val
					}
				}

			case *instructions.RunCommand:
				for _, mount := range instructions.GetMounts(typedCmd) {
					if mount.From != "" {
						from := strings.ToLower(mount.From)
						if val, ok := nameToIndex[from]; ok {
							mount.From = val
						}
					}
				}
			}
		}
	}
}

func GetDockerTargetStageIndex(dockerStages []instructions.Stage, dockerTargetStage string) (int, error) {
	if dockerTargetStage == "" {
		return len(dockerStages) - 1, nil
	}

	for i, s := range dockerStages {
		if s.Name == dockerTargetStage {
			return i, nil
		}
	}

	return -1, fmt.Errorf("%s is not a valid target build stage", dockerTargetStage)
}
