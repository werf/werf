package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportAfterSetupStage(dimgBaseConfig *config.DimgBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterSetupStage {
	imports := getImports(dimgBaseConfig, &getImportsOptions{After: Setup})
	if len(imports) != 0 {
		return newArtifactImportAfterSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newArtifactImportAfterSetupStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterSetupStage {
	s := &ArtifactImportAfterSetupStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports, ArtifactImportAfterSetup, baseStageOptions)
	return s
}

type ArtifactImportAfterSetupStage struct {
	*ArtifactImportStage
}
