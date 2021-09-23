// +build !linux

package buildah

func NewNativeRootlessBuildah() (Buildah, error) {
	panic("not supported")
}
