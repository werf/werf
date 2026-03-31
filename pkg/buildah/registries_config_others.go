//go:build !linux
// +build !linux

package buildah

import "context"

func GetStandaloneInsecureRegistriesFromConfig(_ context.Context) ([]string, error) {
	return nil, nil
}

func GetInsecureRegistriesFromConfig(_ context.Context) ([]string, error) {
	return nil, nil
}

func GetRegistryMirrorsFromConfig(_ context.Context) ([]string, error) {
	return nil, nil
}
