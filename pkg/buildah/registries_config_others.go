//go:build !linux
// +build !linux

package buildah

// GetInsecureRegistriesFromConfig is a no-op stub for non-Linux platforms.
func GetInsecureRegistriesFromConfig() ([]string, error) {
	return nil, nil
}
