//go:build linux
// +build linux

package buildah

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"

	"github.com/werf/logboek"
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
	var paths []string
	seen := make(map[string]bool)

	addPath := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}

	addPath(os.Getenv("CONTAINERS_REGISTRIES_CONF"))
	if home := os.Getenv("HOME"); home != "" {
		addPath(home + "/.config/containers/registries.conf")
	}
	addPath("/etc/containers/registries.conf")

	return paths
}

var (
	cachedRegistries     []registryConf
	errCachedRegistries  error
	cachedRegistriesOnce sync.Once
)

func loadRegistriesConf(ctx context.Context) ([]registryConf, error) {
	cachedRegistriesOnce.Do(func() {
		cachedRegistries, errCachedRegistries = doLoadRegistriesConf(ctx)
	})
	return cachedRegistries, errCachedRegistries
}

func resetRegistriesConfCache() {
	cachedRegistriesOnce = sync.Once{}
	cachedRegistries = nil
	errCachedRegistries = nil
}

func doLoadRegistriesConf(ctx context.Context) ([]registryConf, error) {
	for _, path := range getRegistriesConfPaths() {
		var regs []registryConf
		found := false

		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) && !os.IsPermission(err) {
				return nil, fmt.Errorf("read %s: %w", path, err)
			}
		} else {
			found = true
			var conf registriesConf
			if _, err := toml.Decode(string(data), &conf); err != nil {
				logboek.Context(ctx).Warn().LogF("WARNING: unable to parse %s: %s\n", path, err)
			} else {
				regs = append(regs, conf.Registries...)
			}
		}

		dir := path + ".d"
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				if found {
					return regs, nil
				}
				continue
			}
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		found = true

		var names []string
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
				continue
			}
			names = append(names, entry.Name())
		}
		sort.Strings(names)

		for _, name := range names {
			dropInPath := filepath.Join(dir, name)
			dropInData, err := os.ReadFile(dropInPath)
			if err != nil {
				if os.IsNotExist(err) || os.IsPermission(err) {
					continue
				}
				return nil, fmt.Errorf("read %s: %w", dropInPath, err)
			}

			var dropIn registriesConf
			if _, err := toml.Decode(string(dropInData), &dropIn); err != nil {
				logboek.Context(ctx).Warn().LogF("WARNING: unable to parse %s: %s\n", dropInPath, err)
				continue
			}
			regs = append(regs, dropIn.Registries...)
		}

		if found {
			return regs, nil
		}
	}

	return nil, nil
}

func GetStandaloneInsecureRegistriesFromConfig(ctx context.Context) ([]string, error) {
	regs, err := loadRegistriesConf(ctx)
	if err != nil {
		return nil, err
	}
	if regs == nil {
		return nil, nil
	}

	var result []string
	seen := make(map[string]bool)

	addIfNew := func(loc string) {
		if loc != "" && !seen[loc] {
			seen[loc] = true
			result = append(result, loc)
		}
	}

	for _, reg := range regs {
		if !reg.Insecure {
			continue
		}

		prefix := reg.Prefix
		if prefix == "" {
			prefix = reg.Location
		}

		// Skip docker.io entries — they are mirrors, not standalone registries.
		if prefix == "docker.io" {
			continue
		}

		loc := reg.Location
		if loc == "" {
			loc = reg.Prefix
		}
		addIfNew(loc)
	}

	return result, nil
}

func GetInsecureRegistriesFromConfig(ctx context.Context) ([]string, error) {
	return GetStandaloneInsecureRegistriesFromConfig(ctx)
}

func GetRegistryMirrorsFromConfig(ctx context.Context) ([]string, error) {
	regs, err := loadRegistriesConf(ctx)
	if err != nil {
		return nil, err
	}
	if regs == nil {
		return nil, nil
	}

	var result []string
	seen := make(map[string]bool)

	addMirror := func(loc string, insecure bool) {
		if loc == "" || seen[loc] {
			return
		}
		seen[loc] = true
		// Preserve insecure mirror semantics for downstream code that keys off http:// mirrors.
		if insecure {
			result = append(result, "http://"+loc)
		} else {
			result = append(result, "https://"+loc)
		}
	}

	for _, reg := range regs {
		prefix := reg.Prefix
		if prefix == "" {
			prefix = reg.Location
		}

		if prefix != "docker.io" {
			continue
		}

		if reg.Location != "" && reg.Location != "docker.io" {
			addMirror(reg.Location, reg.Insecure)
		}

		for _, mirror := range reg.Mirrors {
			addMirror(mirror.Location, mirror.Insecure)
		}
	}

	return result, nil
}
