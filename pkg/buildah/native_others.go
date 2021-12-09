//go:build !linux
// +build !linux

package buildah

func NativeProcessStartupHook() bool {
	panic("not supported")
}

func NewNativeBuildah(commonOpts CommonBuildahOpts, opts NativeModeOpts) (Buildah, error) {
	panic("not supported")
}
