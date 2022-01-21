package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesAfterInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *DependenciesAfterInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Install})
	if len(imports) != 0 {
		return newDependenciesAfterInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newDependenciesAfterInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *DependenciesAfterInstallStage {
	s := &DependenciesAfterInstallStage{}
	s.DependenciesStage = newDependenciesStage(imports, DependenciesAfterInstall, baseStageOptions)
	return s
}

type DependenciesAfterInstallStage struct {
	*DependenciesStage
}
