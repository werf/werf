package config

type Image struct {
	*ImageBase
	Docker *Docker
}

func (c *Image) ImageTree() (tree []ImageInterface) {
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

func (c *Image) relatedImages() (relatedImages []ImageInterface) {
	relatedImages = append(relatedImages, c)
	if c.FromImage != nil {
		relatedImages = append(relatedImages, c.FromImage.relatedImages()...)
	}
	if c.FromImageArtifact != nil {
		relatedImages = append(relatedImages, c.FromImageArtifact.relatedImages()...)
	}
	return
}

func (c *Image) lastLayerOrSelf() ImageInterface {
	if c.FromImage != nil {
		return c.FromImage.lastLayerOrSelf()
	}
	if c.FromImageArtifact != nil {
		return c.FromImageArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *Image) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("cannot use shell and ansible builders at the same time!", nil, c.ImageBase.raw.doc)
	}

	return nil
}
