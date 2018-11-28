package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportAfterSetupStage(dimgBaseConfig *config.DimgBase) Interface {
	imports := getImports(dimgBaseConfig, &getImportsOptions{After: "setup"})
	if len(imports) != 0 {
		return newArtifactImportAfterSetupStage(imports)
	}

	return nil
}

func newArtifactImportAfterSetupStage(imports []*config.ArtifactImport) *ArtifactImportAfterSetupStage {
	s := &ArtifactImportAfterSetupStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports)
	return s
}

type ArtifactImportAfterSetupStage struct {
	*ArtifactImportStage
}

func (s *ArtifactImportAfterSetupStage) Name() StageName {
	return ArtifactImportAfterSetup
}
