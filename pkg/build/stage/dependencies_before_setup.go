package stage

import "github.com/werf/werf/pkg/config"

func GenerateDependenciesBeforeSetupStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *DependenciesBeforeSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Setup})
	if len(imports) != 0 {
		return newDependenciesBeforeSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newDependenciesBeforeSetupStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *DependenciesBeforeSetupStage {
	s := &DependenciesBeforeSetupStage{}
	s.DependenciesStage = newDependenciesStage(imports, DependenciesBeforeSetup, baseStageOptions)
	return s
}

type DependenciesBeforeSetupStage struct {
	*DependenciesStage
}
