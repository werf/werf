package config

type DimgArtifact struct {
	*DimgBase
}

func (c *DimgArtifact) DimgTree() (tree []DimgInterface) {
	if c.FromDimg != nil {
		tree = append(tree, c.FromDimg.DimgTree()...)
	}
	if c.FromDimgArtifact != nil {
		tree = append(tree, c.FromDimgArtifact.DimgTree()...)
	}

	for _, importElm := range c.Import {
		tree = append(tree, importElm.ArtifactDimg.DimgTree()...)
	}

	tree = append(tree, c)

	return
}

func (c *DimgArtifact) relatedDimgs() (relatedDimgs []DimgInterface) {
	relatedDimgs = append(relatedDimgs, c)
	if c.FromDimg != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimg.relatedDimgs()...)
	}
	if c.FromDimgArtifact != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimgArtifact.relatedDimgs()...)
	}
	return
}

func (c *DimgArtifact) lastLayerOrSelf() DimgInterface {
	if c.FromDimg != nil {
		return c.FromDimg.lastLayerOrSelf()
	}
	if c.FromDimgArtifact != nil {
		return c.FromDimgArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *DimgArtifact) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("cannot use shell and ansible builders at the same time!", nil, c.DimgBase.raw.doc)
	}

	return nil
}
