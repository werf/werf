package stage

import "github.com/flant/werf/pkg/config"

func GenerateArtifactImportBeforeInstallStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeInstallStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Install})
	if len(imports) != 0 {
		return newArtifactImportBeforeInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newArtifactImportBeforeInstallStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeInstallStage {
	s := &ArtifactImportBeforeInstallStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports, ArtifactImportBeforeInstall, baseStageOptions)
	return s
}

type ArtifactImportBeforeInstallStage struct {
	*ArtifactImportStage
}
