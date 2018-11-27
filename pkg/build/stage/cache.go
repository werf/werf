package stage

type Cache interface {
	GetDimg(name string) *Dimg
	GetImage(imageName string) Image
}
