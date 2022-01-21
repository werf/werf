package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesAfterSetupStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *DependenciesAfterSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Setup})
	if len(imports) != 0 {
		return newDependenciesAfterSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newDependenciesAfterSetupStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *DependenciesAfterSetupStage {
	s := &DependenciesAfterSetupStage{}
	s.DependenciesStage = newDependenciesStage(imports, DependenciesAfterSetup, baseStageOptions)
	return s
}

type DependenciesAfterSetupStage struct {
	*DependenciesStage
}
