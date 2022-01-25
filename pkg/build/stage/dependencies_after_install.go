package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesAfterInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *DependenciesAfterInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Install})
	dependencies := getDependencies(imageBaseConfig, &getImportsOptions{After: Install})
	if len(imports)+len(dependencies) > 0 {
		return newDependenciesAfterInstallStage(imports, dependencies, baseStageOptions)
	}

	return nil
}

func newDependenciesAfterInstallStage(imports []*config.Import, dependencies []*config.Dependency, baseStageOptions *NewBaseStageOptions) *DependenciesAfterInstallStage {
	s := &DependenciesAfterInstallStage{}
	s.DependenciesStage = newDependenciesStage(imports, dependencies, DependenciesAfterInstall, baseStageOptions)
	return s
}

type DependenciesAfterInstallStage struct {
	*DependenciesStage
}
