package stage

import "github.com/flant/dapp/pkg/config"

func GenerateArtifactImportBeforeInstallStage(dimgBaseConfig *config.DimgBase) Interface {
	imports := getImports(dimgBaseConfig, &getImportsOptions{Before: "install"})
	if len(imports) != 0 {
		return newArtifactImportBeforeInstallStage(imports)
	}

	return nil
}

func newArtifactImportBeforeInstallStage(imports []*config.ArtifactImport) *ArtifactImportBeforeInstallStage {
	s := &ArtifactImportBeforeInstallStage{}
	s.ArtifactImportStage = newArtifactImportStage(imports)
	return s
}

type ArtifactImportBeforeInstallStage struct {
	*ArtifactImportStage
}

func (s *ArtifactImportBeforeInstallStage) Name() string {
	return "before_install_artifact"
}
