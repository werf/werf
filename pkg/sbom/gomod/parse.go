package gomod

import (
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

type GoModInfo struct {
	ModulePath          string
	LocalReplaceTargets []string // Old.Path values (e.g. "example.com/mylib")
	LocalReplacePaths   []string // New.Path values (e.g. "./mylib") — Syft may use these as component names
}

func ParseLocalReplaces(goModContent []byte) (*GoModInfo, error) {
	mod, err := modfile.Parse("go.mod", goModContent, nil)
	if err != nil {
		return nil, fmt.Errorf("sbom: parse go.mod: %w", err)
	}

	if mod.Module == nil {
		return nil, fmt.Errorf("sbom: missing module directive")
	}

	info := &GoModInfo{
		ModulePath: mod.Module.Mod.Path,
	}

	for _, replace := range mod.Replace {
		newPath := replace.New.Path
		if strings.HasPrefix(newPath, "./") || strings.HasPrefix(newPath, "../") {
			info.LocalReplaceTargets = append(info.LocalReplaceTargets, replace.Old.Path)
			info.LocalReplacePaths = append(info.LocalReplacePaths, newPath)
			continue
		}

		return nil, fmt.Errorf("sbom: non-local replace for module %q: %s", replace.Old.Path, newPath)
	}

	return info, nil
}
