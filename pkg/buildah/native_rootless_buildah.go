// +build !linux

package buildah

func InitNativeRootlessProcess() (bool, error) {
	panic("not supported")
}

func NewNativeRootlessBuildah() (Buildah, error) {
	panic("not supported")
}
