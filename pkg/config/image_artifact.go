package config

type ImageArtifact struct {
	*ImageBase
}

func (c *ImageArtifact) IsArtifact() bool {
	return true
}

func (c *ImageArtifact) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
