package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesBeforeInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *DependenciesBeforeInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Install})
	if len(imports) != 0 {
		return newDependenciesBeforeInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newDependenciesBeforeInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *DependenciesBeforeInstallStage {
	s := &DependenciesBeforeInstallStage{}
	s.DependenciesStage = newDependenciesStage(imports, DependenciesBeforeInstall, baseStageOptions)
	return s
}

type DependenciesBeforeInstallStage struct {
	*DependenciesStage
}
