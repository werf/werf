package config

import (
	"fmt"
)

type Import struct {
	*ArtifactExport
	ImageName    string
	ArtifactName string
	Before       string
	After        string

	Image         *Image
	ImageArtifact *ImageArtifact

	raw *rawImport
}

func (c *Import) GetRaw() interface{} {
	return c.raw
}

func (c *Import) validate() error {
	if err := c.ArtifactExport.validate(); err != nil {
		return err
	}

	if c.ArtifactName == "" && c.ImageName == "" {
		return newDetailedConfigError("artifact name `artifact: NAME` or image name `image: NAME` required for import!", c.raw, c.raw.rawImage.doc)
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

func (c *Import) associateImportImage(images []*Image, artifacts []*ImageArtifact) error {
	if c.ImageName != "" {
		if image := imageByName(images, c.ImageName); image != nil {
			c.Image = image
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such image `%s`!", c.ImageName), c.raw, c.raw.rawImage.doc)
		}
	} else {
		if imageArtifact := artifactByName(artifacts, c.ArtifactName); imageArtifact != nil {
			c.ImageArtifact = imageArtifact
		} else {
			return newDetailedConfigError(fmt.Sprintf("no such artifact `%s`!", c.ArtifactName), c.raw, c.raw.rawImage.doc)
		}
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
