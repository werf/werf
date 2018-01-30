package config

type ArtifactExport struct {
	*ExportBase

	Raw *RawArtifactExport
}

func (c *ArtifactExport) Validate() error {
	return nil
}
