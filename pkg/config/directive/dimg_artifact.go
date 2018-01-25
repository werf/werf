package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type DimgArtifact struct {
	Dimg
	Shell *ShellArtifact
}

func (c *DimgArtifact) Name() string {
	return c.Artifact
}

func (c *DimgArtifact) ValidateDirectives(artifacts []*DimgArtifact) error {
	if err := c.ValidateImports(artifacts); err != nil {
		return nil
	}

	if c.Shell != nil {
		if err := c.Shell.ValidateDirectives(); err != nil {
			return nil
		}
	}

	if c.Docker != nil {
		fmt.Errorf("docker не поддерживается в dimg artifacte") // FIXME
	}

	return nil
}

// TODO
func (c *DimgArtifact) ToRuby(artifacts []*DimgArtifact) ruby_marshal_config.ArtifactDimg {
	rubyArtifactDimg := ruby_marshal_config.ArtifactDimg{}
	rubyArtifactDimg.Name = c.Name()

	if c.Shell != nil {
		rubyArtifactDimg.Shell = c.Shell.ToRuby()
	}

	//if c.Chef != nil {
	//	rubyArtifactDimg.Chef = c.Chef.ToRuby()
	//}

	//if c.Git != nil {
	//	rubyArtifactDimg.GitArtifact = c.Git.ToRuby()
	//}

	//if c.Mount != nil {
	//	rubyArtifactDimg.Mount = c.Mount.ToRuby()
	//}

	for _, importArtifact := range c.Import {
		dimgArtifact := ArtifactByName(artifacts, importArtifact.ArtifactName)
		artifactExport := ruby_marshal_config.ArtifactExport{}
		artifactExport.Config = dimgArtifact.ToRuby(artifacts)
		rubyArtifactDimg.Artifact = append(rubyArtifactDimg.Artifact)
	}

	return rubyArtifactDimg
}
