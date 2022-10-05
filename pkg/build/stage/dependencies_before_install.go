package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesBeforeInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *BaseStageOptions) *DependenciesBeforeInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Install})
	dependencies := getDependencies(imageBaseConfig, &getImportsOptions{Before: Install})
	if len(imports)+len(dependencies) > 0 {
		return newDependenciesBeforeInstallStage(imports, dependencies, baseStageOptions)
	}

	return nil
}

func newDependenciesBeforeInstallStage(imports []*config.Import, dependencies []*config.Dependency, baseStageOptions *BaseStageOptions) *DependenciesBeforeInstallStage {
	s := &DependenciesBeforeInstallStage{}
	s.DependenciesStage = newDependenciesStage(imports, dependencies, DependenciesBeforeInstall, baseStageOptions)
	return s
}

type DependenciesBeforeInstallStage struct {
	*DependenciesStage
}
