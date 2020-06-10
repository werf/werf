package stage

import "github.com/werf/werf/pkg/config"

func GenerateImportsAfterSetupStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *ImportsAfterSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Setup})
	if len(imports) != 0 {
		return newImportsAfterSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newImportsAfterSetupStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportsAfterSetupStage {
	s := &ImportsAfterSetupStage{}
	s.ImportsStage = newImportsStage(imports, ImportsAfterSetup, baseStageOptions)
	return s
}

type ImportsAfterSetupStage struct {
	*ImportsStage
}
