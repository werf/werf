package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportsBeforeInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportsBeforeInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Install})
	if len(imports) != 0 {
		return newImportsBeforeInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newImportsBeforeInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportsBeforeInstallStage {
	s := &ImportsBeforeInstallStage{}
	s.ImportsStage = newImportsStage(imports, ImportsBeforeInstall, baseStageOptions)
	return s
}

type ImportsBeforeInstallStage struct {
	*ImportsStage
}
