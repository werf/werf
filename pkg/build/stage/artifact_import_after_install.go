package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportAfterInstallStage(dimgBaseConfig *config.DimgBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterInstallStage {
	imports := getImports(dimgBaseConfig, &getImportsOptions{After: Install})
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
