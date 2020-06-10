package stage

import "github.com/werf/werf/pkg/config"

func GenerateImportsBeforeInstallStage(imageBaseConfig *config.StapelImageBase, baseStageOptions *NewBaseStageOptions) *ImportsBeforeInstallStage {
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
