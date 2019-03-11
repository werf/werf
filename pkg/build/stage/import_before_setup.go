package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportBeforeSetupStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportBeforeSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Setup})
	if len(imports) != 0 {
		return newImportBeforeSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newImportBeforeSetupStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportBeforeSetupStage {
	s := &ImportBeforeSetupStage{}
	s.ImportStage = newImportStage(imports, ImportBeforeSetup, baseStageOptions)
	return s
}

type ImportBeforeSetupStage struct {
	*ImportStage
}
