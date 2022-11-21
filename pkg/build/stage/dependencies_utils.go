package stage

import (
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/image"
)

func GetDependenciesArgsKeys(dependencies []*config.Dependency) (res []string) {
	for _, dep := range dependencies {
		for _, imp := range dep.Imports {
			res = append(res, imp.TargetBuildArg)
		}
	}
	return
}

func ResolveDependenciesArgs(dependencies []*config.Dependency, c Conveyor) map[string]string {
	resolved := make(map[string]string)

	for _, dep := range dependencies {
		depImageName := c.GetImageNameForLastImageStage(dep.ImageName)
		depImageID := c.GetImageIDForLastImageStage(dep.ImageName)
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
			}
		}
	}

	return resolved
}
