package image

type ImagesSets [][]*Image

func NewImagesSetsBuilder() *ImagesSetsBuilder {
	return &ImagesSetsBuilder{
		curSetInd:  0,
		nextSetInd: 1,
	}
}

type ImagesSetsBuilder struct {
	allImages  []*Image
	imagesSets ImagesSets
	curSetInd  int
	nextSetInd int
}

func (is *ImagesSetsBuilder) GetImagesSets() ImagesSets {
	return is.imagesSets
}

func (is *ImagesSetsBuilder) GetAllImages() []*Image {
	return is.allImages
}

func (is *ImagesSetsBuilder) MergeImagesSets(newImagesSets [][]*Image) {
	for newSetInd, newSet := range newImagesSets {
		targetSetInd := is.curSetInd + newSetInd
		targetSet := is.getImagesByInd(targetSetInd)
		is.setImagesByInd(targetSetInd, append(targetSet, newSet...))

		for _, img := range newSet {
			is.allImages = append(is.allImages, img)
		}
	}

	nextSetInd := is.curSetInd + len(newImagesSets)
	if nextSetInd > is.nextSetInd {
		is.nextSetInd = nextSetInd
	}
}

func (is *ImagesSetsBuilder) Next() {
	is.curSetInd = is.nextSetInd
}

func (is *ImagesSetsBuilder) getImagesByInd(ind int) []*Image {
	if ind >= len(is.imagesSets) {
		return nil
	}
	return is.imagesSets[ind]
}

func (is *ImagesSetsBuilder) setImagesByInd(ind int, set []*Image) {
	if ind >= len(is.imagesSets) {
		is.imagesSets = append(is.imagesSets, make([][]*Image, ind+1-len(is.imagesSets))...)
	}
	if set != nil {
		is.imagesSets[ind] = set
	}
}
