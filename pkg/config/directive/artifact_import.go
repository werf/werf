package config

import (
	"fmt"
	"github.com/flant/dapp/pkg/config/ruby_marshal_config"
)

type ArtifactImport struct {
	*ExportBase
	ArtifactName string
	ArtifactDimg *DimgArtifact
	Before       string
	After        string
}

func (c *ArtifactImport) Validate() error {
	if err := c.ExportBase.Validate(); err != nil {
		return err
	}

	if c.ArtifactName == "" {
		return fmt.Errorf("имя артефакта обязательно!") // FIXME
	} else if c.Before != "" && c.After != "" {
		return fmt.Errorf("артефакт не может иметь несколько связанных стадий!") // FIXME
	} else if c.Before == "" && c.After == "" {
		return fmt.Errorf("артефакт должен иметь связанную стадию!") // FIXME
	} else if c.Before != "" && checkInvalidRelation(c.Before) {
		return fmt.Errorf("артефакт имеет некорректную связанную стадию (before)!") // FIXME
	} else if c.After != "" && checkInvalidRelation(c.After) {
		return fmt.Errorf("артефакт должен иметь связанную стадию (after)! %s") // FIXME
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
		return fmt.Errorf("артефакт из импорта не найден!") // FIXME
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
	artifactExport.After = c.After
	artifactExport.Before = c.Before
	return artifactExport
}
