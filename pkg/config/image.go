package config

type Image struct {
	*ImageBase
	Docker *Docker
}

func (c *Image) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("can not use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
