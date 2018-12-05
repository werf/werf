package config

type Dimg struct {
	*DimgBase
	Docker *Docker
}

func (c *Dimg) DimgTree() (tree []DimgInterface) {
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

func (c *Dimg) relatedDimgs() (relatedDimgs []DimgInterface) {
	relatedDimgs = append(relatedDimgs, c)
	if c.FromDimg != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimg.relatedDimgs()...)
	}
	if c.FromDimgArtifact != nil {
		relatedDimgs = append(relatedDimgs, c.FromDimgArtifact.relatedDimgs()...)
	}
	return
}

func (c *Dimg) lastLayerOrSelf() DimgInterface {
	if c.FromDimg != nil {
		return c.FromDimg.lastLayerOrSelf()
	}
	if c.FromDimgArtifact != nil {
		return c.FromDimgArtifact.lastLayerOrSelf()
	}
	return c
}

func (c *Dimg) validate() error {
	if !oneOrNone([]bool{c.Shell != nil, c.Ansible != nil}) {
		return newDetailedConfigError("Cannot use shell and ansible builders at the same time!", nil, c.DimgBase.raw.doc)
	}

	return nil
}
