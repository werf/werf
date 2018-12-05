package config

import (
	"fmt"
)

type ArtifactImport struct {
	*ArtifactExport
	ArtifactName string
	Before       string
	After        string

	ArtifactDimg *DimgArtifact

	raw *rawArtifactImport
}

func (c *ArtifactImport) validate() error {
	if err := c.ArtifactExport.validate(); err != nil {
		return err
	}

	if c.ArtifactName == "" {
		return newDetailedConfigError("Artifact name `artifact: NAME` required for import!", c.raw, c.raw.rawDimg.doc)
	} else if c.Before != "" && c.After != "" {
		return newDetailedConfigError("Specify only one artifact stage using `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawDimg.doc)
	} else if c.Before == "" && c.After == "" {
		return newDetailedConfigError("Artifact stage is not specified with `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawDimg.doc)
	} else if c.Before != "" && checkInvalidRelation(c.Before) {
		return newDetailedConfigError(fmt.Sprintf("Invalid artifact stage `before: %s` for import: expected install or setup!", c.Before), c.raw, c.raw.rawDimg.doc)
	} else if c.After != "" && checkInvalidRelation(c.After) {
		return newDetailedConfigError(fmt.Sprintf("Invalid artifact stage `after: %s` for import: expected install or setup!", c.After), c.raw, c.raw.rawDimg.doc)
	}
	return nil
}

func checkInvalidRelation(rel string) bool {
	return !(rel == "install" || rel == "setup")
}

func (c *ArtifactImport) associateArtifact(artifacts []*DimgArtifact) error {
	if artifactDimg := artifactByName(artifacts, c.ArtifactName); artifactDimg != nil {
		c.ArtifactDimg = artifactDimg
	} else {
		return newDetailedConfigError(fmt.Sprintf("No such artifact `%s`!", c.ArtifactName), c.raw, c.raw.rawDimg.doc)
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
