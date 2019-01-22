package config

import (
	"fmt"
)

type ArtifactImport struct {
	*ArtifactExport
	ArtifactName string
	Before       string
	After        string

	ImageArtifact *ImageArtifact

	raw *rawArtifactImport
}

func (c *ArtifactImport) GetRaw() interface{} {
	return c.raw
}

func (c *ArtifactImport) validate() error {
	if err := c.ArtifactExport.validate(); err != nil {
		return err
	}

	if c.ArtifactName == "" {
		return newDetailedConfigError("artifact name `artifact: NAME` required for import!", c.raw, c.raw.rawImage.doc)
	} else if c.Before != "" && c.After != "" {
		return newDetailedConfigError("specify only one artifact stage using `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawImage.doc)
	} else if c.Before == "" && c.After == "" {
		return newDetailedConfigError("artifact stage is not specified with `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawImage.doc)
	} else if c.Before != "" && checkInvalidRelation(c.Before) {
		return newDetailedConfigError(fmt.Sprintf("invalid artifact stage `before: %s` for import: expected install or setup!", c.Before), c.raw, c.raw.rawImage.doc)
	} else if c.After != "" && checkInvalidRelation(c.After) {
		return newDetailedConfigError(fmt.Sprintf("invalid artifact stage `after: %s` for import: expected install or setup!", c.After), c.raw, c.raw.rawImage.doc)
	}
	return nil
}

func checkInvalidRelation(rel string) bool {
	return !(rel == "install" || rel == "setup")
}

func (c *ArtifactImport) associateArtifact(artifacts []*ImageArtifact) error {
	if imageArtifact := artifactByName(artifacts, c.ArtifactName); imageArtifact != nil {
		c.ImageArtifact = imageArtifact
	} else {
		return newDetailedConfigError(fmt.Sprintf("no such artifact `%s`!", c.ArtifactName), c.raw, c.raw.rawImage.doc)
	}
	return nil
}

func artifactByName(artifacts []*ImageArtifact, name string) *ImageArtifact {
	for _, artifact := range artifacts {
		if artifact.Name == name {
			return artifact
		}
	}
	return nil
}
