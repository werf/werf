package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportAfterInstallStage(dimgBaseConfig *config.DimgBase) Interface {
	imports := getImports(dimgBaseConfig, &getImportsOptions{After: "install"})
	if len(imports) != 0 {
		return newArtifactImportAfterInstallStage(imports)
	}

	return nil
}

func newArtifactImportAfterInstallStage(imports []*config.ArtifactImport) *ArtifactImportAfterInstallStage {
	s := &ArtifactImportAfterInstallStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports)
	return s
}

type ArtifactImportAfterInstallStage struct {
	*ArtifactImportStage
}

func (s *ArtifactImportAfterInstallStage) Name() StageName {
	return ArtifactImportAfterInstall
}
