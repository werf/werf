package stage

import (
	"fmt"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/image"
)

func GetDependenciesArgsKeys(dependencies []*config.Dependency) (res []string) {
	for _, dep := range dependencies {
		for _, imp := range dep.Imports {
			res = append(res, imp.TargetBuildArg)
		}
	}
	return
}

func ResolveDependenciesArgs(targetPlatform string, dependencies []*config.Dependency, c Conveyor) map[string]string {
	resolved := make(map[string]string)

	for _, dep := range dependencies {
		depImageName := c.GetImageNameForLastImageStage(targetPlatform, dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(targetPlatform, dep.ImageName)
		depImageDigest := c.GetImageDigestForLastImageStage(targetPlatform, dep.ImageName)
		depImageRepo, depImageTag := image.ParseRepositoryAndTag(depImageName)

		for _, imp := range dep.Imports {
			switch imp.Type {
			case config.ImageRepoImport:
				resolved[imp.TargetBuildArg] = depImageRepo
			case config.ImageTagImport:
				resolved[imp.TargetBuildArg] = depImageTag
			case config.ImageNameImport:
				resolved[imp.TargetBuildArg] = depImageName
			case config.ImageIDImport:
				resolved[imp.TargetBuildArg] = depImageID
			case config.ImageDigestImport:
				resolved[imp.TargetBuildArg] = depImageDigest
			}
		}
	}

	return resolved
}

func dependencyLabelKey(depStageID string) string {
	id := depStageID
	if len(depStageID) > 8 {
		id = depStageID[:8]
	}
	return fmt.Sprintf("%s%s", image.WerfDependencySourceStageIDLabelPrefix, id)
}
