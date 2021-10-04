// +build !linux

package buildah

func NativeRootlessProcessStartupHook() bool {
	panic("not supported")
}

func NewNativeRootlessBuildah(commonOpts CommonBuildahOpts, opts NativeRootlessModeOpts) (Buildah, error) {
	panic("not supported")
}
