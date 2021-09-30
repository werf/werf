// +build !linux

package buildah

func InitNativeRootlessProcess() (bool, error) {
	panic("not supported")
}

func NewNativeRootlessBuildah(commonOpts CommonBuildahOpts, opts NativeRootlessModeOpts) (Buildah, error) {
	panic("not supported")
}
