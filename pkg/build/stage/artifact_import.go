package stage

import (
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
)

type getImportsOptions struct {
	Before string
	After  string
}

func getImports(dimgBaseConfig *config.DimgBase, options *getImportsOptions) []*config.ArtifactImport {
	var imports []*config.ArtifactImport
	for _, elm := range dimgBaseConfig.Import {
		if options.Before != "" && elm.Before != "" && elm.Before == options.Before {
			imports = append(imports, elm)
		} else if options.After != "" && elm.After != "" && elm.After == options.After {
			imports = append(imports, elm)
		}
	}

	return imports
}

func newArtifactImportStage(imports []*config.ArtifactImport) *ArtifactImportStage {
	s := &ArtifactImportStage{}
	s.imports = imports
	s.BaseStage = newBaseStage()

	return s
}

type ArtifactImportStage struct {
	*BaseStage

	imports []*config.ArtifactImport
}

func (s *ArtifactImportStage) GetDependencies(c Cache) string {
	var args []string

	for _, elm := range s.imports {
		args = append(args, c.GetDimg(elm.ArtifactName).LatestStage().GetSignature())
		args = append(args, elm.Add, elm.To)
		args = append(args, elm.Group, elm.Owner)
		args = append(args, elm.IncludePaths...)
		args = append(args, elm.ExcludePaths...)
	}

	return util.Sha256Hash(args...)
}
