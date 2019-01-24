package config

type ImageArtifact struct {
	*ImageBase
}

func (c *ImageArtifact) IsArtifact() bool {
	return true
}

func (c *ImageArtifact) ImageTree() (tree []ImageInterface) {
	if c.FromImage != nil {
		tree = append(tree, c.FromImage.ImageTree()...)
	}
	if c.FromImageArtifact != nil {
		tree = append(tree, c.FromImageArtifact.ImageTree()...)
	}

	for _, importElm := range c.Import {
		tree = append(tree, importElm.ImageArtifact.ImageTree()...)
	}

	tree = append(tree, c)

	return
}

func (c *ImageArtifact) relatedImages() (relatedImages []ImageInterface) {
	relatedImages = append(relatedImages, c)
	if c.FromImage != nil {
		relatedImages = append(relatedImages, c.FromImage.relatedImages()...)
	}
	if c.FromImageArtifact != nil {
		relatedImages = append(relatedImages, c.FromImageArtifact.relatedImages()...)
	}
	return
}

func (c *ImageArtifact) lastLayerOrSelf() ImageInterface {
	if c.FromImage != nil {
		return c.FromImage.lastLayerOrSelf()
	}
	if c.FromImageArtifact != nil {
		return c.FromImageArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *ImageArtifact) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("cannot use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
