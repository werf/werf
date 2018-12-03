package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

type getImportsOptions struct {
	Before StageName
	After  StageName
}

func getImports(dimgBaseConfig *config.DimgBase, options *getImportsOptions) []*config.ArtifactImport {
	var imports []*config.ArtifactImport
	for _, elm := range dimgBaseConfig.Import {
		if options.Before != "" && elm.Before != "" && elm.Before == string(options.Before) {
			imports = append(imports, elm)
		} else if options.After != "" && elm.After != "" && elm.After == string(options.After) {
			imports = append(imports, elm)
		}
	}

	return imports
}

func newArtifactImportStage(imports []*config.ArtifactImport, baseStageOptions *NewBaseStageOptions) *ArtifactImportStage {
	s := &ArtifactImportStage{}
	s.imports = imports
	s.BaseStage = newBaseStage(baseStageOptions)
	return s
}

type ArtifactImportStage struct {
	*BaseStage

	imports []*config.ArtifactImport
}

func (s *ArtifactImportStage) GetDependencies(c Conveyor, _ image.Image) (string, error) {
	var args []string

	for _, elm := range s.imports {
		args = append(args, c.GetDimgSignature(elm.ArtifactName))
		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)
	}

	return util.Sha256Hash(args...), nil
}
