package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportAfterInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportAfterInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Install})
	if len(imports) != 0 {
		return newImportAfterInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newImportAfterInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportAfterInstallStage {
	s := &ImportAfterInstallStage{}
	s.ImportStage = newImportStage(imports, ImportAfterInstall, baseStageOptions)
	return s
}

type ImportAfterInstallStage struct {
	*ImportStage
}
