package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportBeforeInstallStage(dimgBaseConfig *config.DimgBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeInstallStage {
	imports := getImports(dimgBaseConfig, &getImportsOptions{Before: Install})
	if len(imports) != 0 {
		return newArtifactImportBeforeInstallStage(imports, baseStageOptions)
	}

	return nil
}

func newArtifactImportBeforeInstallStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeInstallStage {
	s := &ArtifactImportBeforeInstallStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports, baseStageOptions)
	return s
}

type ArtifactImportBeforeInstallStage struct {
	*ArtifactImportStage
}

func (s *ArtifactImportBeforeInstallStage) Name() StageName {
	return ArtifactImportBeforeInstall
}
