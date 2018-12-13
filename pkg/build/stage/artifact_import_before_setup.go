package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportBeforeSetupStage(dimgBaseConfig *config.DimgBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeSetupStage {
	imports := getImports(dimgBaseConfig, &getImportsOptions{Before: Setup})
	if len(imports) != 0 {
		return newArtifactImportBeforeSetupStage(imports, baseStageOptions)
	}

	return nil
}

func newArtifactImportBeforeSetupStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeSetupStage {
	s := &ArtifactImportBeforeSetupStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports, ArtifactImportBeforeSetup, baseStageOptions)
	return s
}

type ArtifactImportBeforeSetupStage struct {
	*ArtifactImportStage
}
