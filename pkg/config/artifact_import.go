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
		return fmt.Errorf("Artifact name required!\n\n", DumpConfigSection(c.Raw)) // FIXME
	} else if c.Before != "" && c.After != "" {
		return fmt.Errorf("Specify only one artifact stage using `before: <stage>` or `after: <stage>`!") // FIXME
	} else if c.Before == "" && c.After == "" {
		return fmt.Errorf("Artifact stage is not specified with `before: STAGE` or `after: STAGE` for import!\n\n%s\n%s", DumpConfigSection(c.Raw), DumpConfigDoc(c.Raw.RawDimg.Doc))
	} else if c.Before != "" && checkInvalidRelation(c.Before) {
		return fmt.Errorf("Invalid artifact stage `before: %s`: expected install or setup!", c.Before) // FIXME
	} else if c.After != "" && checkInvalidRelation(c.After) {
		return fmt.Errorf("Invalid artifact stage `after: %s`: expected install or setup!", c.After) // FIXME
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
		return fmt.Errorf("No such artifact `%s`!", c.ArtifactName) // FIXME
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
