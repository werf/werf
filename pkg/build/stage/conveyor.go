package stage

type Conveyor interface {
	GetDimgSignature(dimgName string) string
	// GetDimg(name string) *Dimg
	// GetImage(imageName string) Image
}
