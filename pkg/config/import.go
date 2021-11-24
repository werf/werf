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
	Stage        string

	raw *rawImport
}

func (c *Import) GetRaw() interface{} {
	return c.raw
}

func (c *Import) validate() error {
	if err := c.ArtifactExport.validate(); err != nil {
		return err
	}

	switch {
	case c.ArtifactName == "" && c.ImageName == "":
		return newDetailedConfigError("artifact name `artifact: NAME` or image name `image: NAME` required for import!", c.raw, c.raw.rawStapelImage.doc)
	case c.ArtifactName != "" && c.ImageName != "":
		return newDetailedConfigError("specify only one artifact name using `artifact: NAME` or image name using `image: NAME` for import!", c.raw, c.raw.rawStapelImage.doc)
	case c.Before != "" && c.After != "":
		return newDetailedConfigError("specify only one artifact stage using `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawStapelImage.doc)
	case c.Before == "" && c.After == "":
		return newDetailedConfigError("artifact stage is not specified with `before: install|setup` or `after: install|setup` for import!", c.raw, c.raw.rawStapelImage.doc)
	case c.Before != "" && checkInvalidRelation(c.Before):
		return newDetailedConfigError(fmt.Sprintf("invalid artifact stage `before: %s` for import: expected install or setup!", c.Before), c.raw, c.raw.rawStapelImage.doc)
	case c.After != "" && checkInvalidRelation(c.After):
		return newDetailedConfigError(fmt.Sprintf("invalid artifact stage `after: %s` for import: expected install or setup!", c.After), c.raw, c.raw.rawStapelImage.doc)
	case c.Stage != "" && checkInvalidStage(c.Stage):
		return newDetailedConfigError(fmt.Sprintf("invalid stage `stage: %s` for import: expected beforeInstall, install, beforeSetup or setup", c.Stage), c.raw, c.raw.rawStapelImage.doc)
	default:
		return nil
	}
}

func checkInvalidRelation(rel string) bool {
	return !(rel == "install" || rel == "setup")
}

func checkInvalidStage(stage string) bool {
	return !(stage == "beforeInstall" || stage == "install" || stage == "beforeSetup" || stage == "setup")
}
