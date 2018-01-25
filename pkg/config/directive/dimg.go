package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type Dimg struct {
	DimgBase
	Shell *ShellDimg
}

func (c *Dimg) Names() []string {
	name, typeDimg := c.Dimg.(string)
	names, typeDimgArray := c.Dimg.([]string)

	if typeDimg {
		return []string{name}
	} else if typeDimgArray {
		return names
	}
	return nil
}

func (c *Dimg) ValidateDirectives(artifacts []*DimgArtifact) error { // FIXME: переменовать: вызов происходит просле разбиения DimgBase на Dimg и DimgArtifact
	if err := c.ValidateImports(artifacts); err != nil {
		return err
	}

	if err := c.Shell.ValidateDirectives(); err != nil {
		return err
	}
	return nil
}

func (c *Dimg) ValidateImports(artifacts []*DimgArtifact) error {
	if c.ValidateImports(artifacts) != nil {
		for _, importArtifact := range c.Import {
			if ArtifactByName(artifacts, importArtifact.ArtifactName) == nil {
				return fmt.Errorf("нет соответствующего артефакта для импорта!")
			}
		}
	}
	return nil
}

// TODO
func (c *Dimg) ToRuby(artifacts []*DimgArtifact) []ruby_marshal_config.Dimg {
	var rubyDimgs []ruby_marshal_config.Dimg

	for _, dimgName := range c.Names() {
		rubyDimg := ruby_marshal_config.Dimg{}
		rubyDimg.Name = dimgName

		if c.Shell != nil {
			rubyDimg.Shell = c.Shell.ToRuby()
		}

		//if c.Chef != nil {
		//	rubyDimg.Chef = c.Chef.ToRuby()
		//}

		//if c.Docker != nil {
		//	rubyDimg.Docker = c.Docker.ToRuby()
		//}

		//if c.Git != nil {
		//	rubyDimg.GitArtifact = c.Git.ToRuby()
		//}

		//if c.Mount != nil {
		//	rubyDimg.Mount = c.Mount.ToRuby()
		//}

		for _, importArtifact := range c.Import {
			dimgArtifact := ArtifactByName(artifacts, importArtifact.ArtifactName)
			artifactExport := ruby_marshal_config.ArtifactExport{}
			artifactExport.Config = dimgArtifact.ToRuby(artifacts)
			rubyDimg.Artifact = append(rubyDimg.Artifact)
		}

		rubyDimgs = append(rubyDimgs, rubyDimg)
	}

	return rubyDimgs
}
