package config

type StapelImageArtifact struct {
	*StapelImageBase
}

func (c *StapelImageArtifact) IsArtifact() bool {
	return true
}

func (c *StapelImageArtifact) IsFinal() bool {
	return false
}

func (c *StapelImageArtifact) validate() error {
	printArtifactDepricationWarning()

	return nil
}
