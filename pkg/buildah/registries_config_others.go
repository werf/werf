//go:build !linux
// +build !linux

package buildah

func GetInsecureRegistriesFromConfig() ([]string, error) {
	return nil, nil
}

func GetRegistryMirrorsFromConfig() ([]string, error) {
	return nil, nil
}
