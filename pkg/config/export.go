package config

type Export struct {
	*ExportBase

	raw *rawExport
}

func (c *Export) validate() error {
	return c.ExportBase.validate()
}
