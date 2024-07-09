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

	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.StapelImageBase.raw.doc)
	}

	return nil
}
