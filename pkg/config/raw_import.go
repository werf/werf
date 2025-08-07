package config

import (
	"github.com/docker/distribution/reference"
)

type rawImport struct {
	ImageName    string `yaml:"image,omitempty"`
	From         string `yaml:"from,omitempty"`
	ArtifactName string `yaml:"artifact,omitempty"`
	Before       string `yaml:"before,omitempty"`
	After        string `yaml:"after,omitempty"`
	Stage        string `yaml:"stage,omitempty"`

	rawArtifactExport `yaml:",inline"`
	rawStapelImage    *rawStapelImage `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (c *rawImport) configSection() interface{} {
	return c
}

func (c *rawImport) doc() *doc {
	return c.rawStapelImage.doc
}

func (c *rawImport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawStapelImage); ok {
		c.rawStapelImage = parent
	}

	parentStack.Push(c)
	type plain rawImport
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	c.rawArtifactExport.inlinedIntoRaw(c)

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawStapelImage.doc); err != nil {
		return err
	}

	if c.rawArtifactExport.rawExportBase.To == "" {
		c.rawArtifactExport.rawExportBase.To = c.rawArtifactExport.rawExportBase.Add
	}

	return nil
}

func (c *rawImport) toDirective() (imp *Import, err error) {
	imp = &Import{}

	if artifactExport, err := c.rawArtifactExport.toDirective(); err != nil {
		return nil, err
	} else {
		imp.ArtifactExport = artifactExport
	}

	if !oneOrNone([]bool{c.ImageName != "", c.From != ""}) {
		return nil, newDetailedConfigError("specify only `image: NAME` or `from: NAME` for import!", c, c.doc())
	}

	if c.From != "" {
		imp.ImageName = c.From
	} else {
		imp.ImageName = c.ImageName // to deprecate
	}

	if hasTagOrDigest(imp.ImageName) {
		imp.ExternalImage = true
	}

	imp.ArtifactName = c.ArtifactName
	imp.Before = c.Before
	imp.After = c.After
	imp.Stage = c.Stage

	imp.raw = c

	if err = c.validateDirective(imp); err != nil {
		return nil, err
	}

	return imp, nil
}

func (c *rawImport) validateDirective(imp *Import) (err error) {
	if err = imp.validate(); err != nil {
		return err
	}

	return nil
}

func hasTagOrDigest(image string) bool {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return false
	}

	_, isTagged := ref.(reference.Tagged)
	_, isDigested := ref.(reference.Digested)

	return isTagged || isDigested
}
