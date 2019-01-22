package stage

import "github.com/flant/werf/pkg/config"

func GenerateArtifactImportAfterInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Install})
	if len(imports) != 0 {
		return newArtifactImportAfterInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newArtifactImportAfterInstallStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterInstallStage {
	s := &ArtifactImportAfterInstallStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports, ArtifactImportAfterInstall, baseStageOptions)
	return s
}

type ArtifactImportAfterInstallStage struct {
	*ArtifactImportStage
}
