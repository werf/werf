package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ArtifactImport struct {
	*ArtifactExport
	ArtifactName string
	ArtifactDimg *DimgArtifact
	Before       string
	After        string

	Raw *RawArtifactImport
}

func (c *ArtifactImport) Validate() error {
	if err := c.ArtifactExport.Validate(); err != nil {
		return err
	}

	if c.ArtifactName == "" {
		return fmt.Errorf("Artifact name `artifact: NAME` required for import!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	} else if c.Before != "" && c.After != "" {
		return fmt.Errorf("Specify only one artifact stage using `before: install|setup` or `after: install|setup` for import!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	} else if c.Before == "" && c.After == "" {
		return fmt.Errorf("Artifact stage is not specified with `before: install|setup` or `after: install|setup` for import!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	} else if c.Before != "" && checkInvalidRelation(c.Before) {
		return fmt.Errorf("Invalid artifact stage `before: %s` for import: expected install or setup!\n\n%s\n%s", c.Before, DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	} else if c.After != "" && checkInvalidRelation(c.After) {
		return fmt.Errorf("Invalid artifact stage `after: %s` for import: expected install or setup!\n\n%s\n%s", c.After, DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	}
	return nil
}

func checkInvalidRelation(rel string) bool {
	return !(rel == "install" || rel == "setup")
}

func (c *ArtifactImport) AssociateArtifact(artifacts []*DimgArtifact) error {
	if artifactDimg := artifactByName(artifacts, c.ArtifactName); artifactDimg != nil {
		c.ArtifactDimg = artifactDimg
	} else {
		return fmt.Errorf("No such artifact `%s`!\n\n%s\n%s", c.ArtifactName, DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	}
	return nil
}

func artifactByName(artifacts []*DimgArtifact, name string) *DimgArtifact {
	for _, artifact := range artifacts {
		if artifact.Name == name {
			return artifact
		}
	}
	return nil
}

func (c *ArtifactImport) ToRuby() ruby_marshal_config.ArtifactExport {
	artifactExport := ruby_marshal_config.ArtifactExport{}

	if c.ExportBase != nil {
		artifactExport.ArtifactBaseExport = c.ExportBase.ToRuby()
	}
	if c.ArtifactDimg != nil {
		artifactExport.Config = c.ArtifactDimg.ToRuby()
	}

	artifactExport.After = ruby_marshal_config.Symbol(c.After)
	artifactExport.Before = ruby_marshal_config.Symbol(c.Before)
	return artifactExport
}
