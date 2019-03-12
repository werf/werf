package stage

import "github.com/flant/werf/pkg/config"

func GenerateImportsAfterInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ImportsAfterInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Install})
	if len(imports) != 0 {
		return newImportsAfterInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newImportsAfterInstallStage(imports []*config.Import, baseStageOptions *NewBaseStageOptions) *ImportsAfterInstallStage {
	s := &ImportsAfterInstallStage{}
	s.ImportsStage = newImportsStage(imports, ImportsAfterInstall, baseStageOptions)
	return s
}

type ImportsAfterInstallStage struct {
	*ImportsStage
}
