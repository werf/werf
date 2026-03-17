//go:build linux
// +build linux

package buildah

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type registriesConf struct {
	Registries []registryConf `toml:"registry"`
}

type registryConf struct {
	Location string       `toml:"location"`
	Prefix   string       `toml:"prefix"`
	Insecure bool         `toml:"insecure"`
	Mirrors  []mirrorConf `toml:"mirror"`
}

type mirrorConf struct {
	Location string `toml:"location"`
	Insecure bool   `toml:"insecure"`
}

func getRegistriesConfPaths() []string {
	paths := []string{
		"/etc/containers/registries.conf",
	}
	if home := os.Getenv("HOME"); home != "" {
		paths = append([]string{home + "/.config/containers/registries.conf"}, paths...)
	}
	return paths
}

func GetInsecureRegistriesFromConfig() ([]string, error) {
	for _, path := range getRegistriesConfPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var conf registriesConf
		if _, err := toml.Decode(string(data), &conf); err != nil {
			continue
		}

		var result []string
		seen := make(map[string]bool)

		addIfNew := func(loc string) {
			if loc != "" && !seen[loc] {
				seen[loc] = true
				result = append(result, loc)
			}
		}

		for _, reg := range conf.Registries {
			if reg.Insecure {
				loc := reg.Location
				if loc == "" {
					loc = reg.Prefix
				}
				addIfNew(loc)
			}

			for _, mirror := range reg.Mirrors {
				if mirror.Insecure {
					addIfNew(mirror.Location)
				}
			}
		}

		return result, nil
	}

	return nil, nil
}

// GetRegistryMirrorsFromConfig returns docker.io mirrors from registries.conf.
func GetRegistryMirrorsFromConfig() ([]string, error) {
	for _, path := range getRegistriesConfPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		var conf registriesConf
		if _, err := toml.Decode(string(data), &conf); err != nil {
			continue
		}

		var result []string
		seen := make(map[string]bool)

		for _, reg := range conf.Registries {
			loc := reg.Location
			if loc == "" {
				loc = reg.Prefix
			}

			if loc != "docker.io" {
				continue
			}

			for _, mirror := range reg.Mirrors {
				if mirror.Location != "" && !seen[mirror.Location] {
					seen[mirror.Location] = true
					result = append(result, "https://"+mirror.Location)
				}
			}
		}

		return result, nil
	}

	return nil, nil
}
