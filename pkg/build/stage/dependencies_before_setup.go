package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesBeforeSetupStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *BaseStageOptions) *DependenciesBeforeSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Setup})
	dependencies := getDependencies(imageBaseConfig, &getImportsOptions{Before: Setup})
	if len(imports)+len(dependencies) > 0 {
		return newDependenciesBeforeSetupStage(imports, dependencies, baseStageOptions)
	}

	return nil
}

func newDependenciesBeforeSetupStage(imports []*config.Import, dependencies []*config.Dependency, baseStageOptions *BaseStageOptions) *DependenciesBeforeSetupStage {
	s := &DependenciesBeforeSetupStage{}
	s.DependenciesStage = newDependenciesStage(imports, dependencies, DependenciesBeforeSetup, baseStageOptions)
	return s
}

type DependenciesBeforeSetupStage struct {
	*DependenciesStage
}
