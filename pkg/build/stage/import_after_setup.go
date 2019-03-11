package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportAfterSetupStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportAfterSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Setup})
	if len(imports) != 0 {
		return newImportAfterSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newImportAfterSetupStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportAfterSetupStage {
	s := &ImportAfterSetupStage{}
	s.ImportStage = newImportStage(imports, ImportAfterSetup, baseStageOptions)
	return s
}

type ImportAfterSetupStage struct {
	*ImportStage
}
