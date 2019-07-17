package config

type StapelImage struct {
	*StapelImageBase
	Docker *Docker
}

func (c *StapelImage) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.StapelImageBase.raw.doc)
	}

	return nil
}
