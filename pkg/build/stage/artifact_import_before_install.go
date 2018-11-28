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
	s.ArtifactImportBaseStage = newArtifactImportBaseStage(imports)
	return s
}

type ArtifactImportBeforeInstallStage struct {
	*ArtifactImportBaseStage
}

func (s *ArtifactImportBeforeInstallStage) Name() string {
	return "before_install_artifact"
}
