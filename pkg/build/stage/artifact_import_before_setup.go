package stage

import "github.com/flant/werf/pkg/config"

func GenerateArtifactImportBeforeSetupStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportBeforeSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{Before: Setup})
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
