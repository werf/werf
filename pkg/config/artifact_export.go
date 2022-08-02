package config

type ArtifactExport struct {
	*ExportBase

	raw *rawArtifactExport
}

func (c *ArtifactExport) validate() error {
	return c.ExportBase.validate()
}
