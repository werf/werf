package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesAfterSetupStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *BaseStageOptions) *DependenciesAfterSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Setup})
	dependencies := getDependencies(imageBaseConfig, &getImportsOptions{After: Setup})
	if len(imports)+len(dependencies) > 0 {
		return newDependenciesAfterSetupStage(imports, dependencies, baseStageOptions)
	}

	return nil
}

func newDependenciesAfterSetupStage(imports []*config.Import, dependencies []*config.Dependency, baseStageOptions *BaseStageOptions) *DependenciesAfterSetupStage {
	s := &DependenciesAfterSetupStage{}
	s.DependenciesStage = newDependenciesStage(imports, dependencies, DependenciesAfterSetup, baseStageOptions)
	return s
}

type DependenciesAfterSetupStage struct {
	*DependenciesStage
}
