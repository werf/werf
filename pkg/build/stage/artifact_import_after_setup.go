package stage

import "github.com/flant/werf/pkg/config"

func GenerateArtifactImportAfterSetupStage(imageBaseConfig *config.ImageBase, baseStageOptions *NewBaseStageOptions) *ArtifactImportAfterSetupStage {
	imports := getImports(imageBaseConfig, &getImportsOptions{After: Setup})
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
