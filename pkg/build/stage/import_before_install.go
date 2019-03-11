package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportBeforeInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportBeforeInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Install})
	if len(imports) != 0 {
		return newImportBeforeInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newImportBeforeInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportBeforeInstallStage {
	s := &ImportBeforeInstallStage{}
	s.ImportStage = newImportStage(imports, ImportBeforeInstall, baseStageOptions)
	return s
}

type ImportBeforeInstallStage struct {
	*ImportStage
}
