package config

type Cookbook struct {
	Name              string
	VersionConstraint string
	Path              string
	Fields            map[string]interface{}

	Raw *RawCookbook
}

func (c *Cookbook) Validate() error {
	return nil
}
