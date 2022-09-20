package build

type ImagesToProcess struct {
	OnlyImages    []string
	WithoutImages bool
}

func NewImagesToProcess(onlyImages []string, withoutImages bool) ImagesToProcess {
	return ImagesToProcess{OnlyImages: onlyImages, WithoutImages: withoutImages}
}
