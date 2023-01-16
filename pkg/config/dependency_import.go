package config

import (
	"fmt"
	"strings"
)

type DependencyImport struct {
	Type           DependencyImportType
	TargetBuildArg string
	TargetEnv      string

	raw *rawDependencyImport
}

func (i *DependencyImport) validate() error {
	switch i.Type {
	case ImageNameImport, ImageTagImport, ImageRepoImport, ImageIDImport, ImageDigestImport:
	default:
		return newDetailedConfigError(fmt.Sprintf("invalid `type: %s` for dependency import, expected one of: %s", i.Type, strings.Join([]string{string(ImageNameImport), string(ImageTagImport), string(ImageRepoImport), string(ImageIDImport), string(ImageDigestImport)}, ", ")), i.raw, i.raw.rawDependency.doc())
	}

	switch imgType := i.raw.rawDependency.imageType(); imgType {
	case dependencyImageTypeStapel:
		switch {
		case i.TargetEnv == "":
			return newDetailedConfigError("targetEnv directive cannot be empty for a Stapel image", i.raw, i.raw.rawDependency.doc())
		case i.TargetBuildArg != "":
			return newDetailedConfigError("targetBuildArg directive cannot be used for a Stapel image", i.raw, i.raw.rawDependency.doc())
		}
	case dependencyImageTypeDockerfile:
		switch {
		case i.TargetBuildArg == "":
			return newDetailedConfigError("targetBuildArg directive cannot be empty for a Dockerfile image", i.raw, i.raw.rawDependency.doc())
		case i.TargetEnv != "":
			return newDetailedConfigError("targetEnv directive cannot be used for a Dockerfile image", i.raw, i.raw.rawDependency.doc())
		}
	default:
		panic(fmt.Sprintf("unexpected dependencyImageType: %s", imgType))
	}

	return nil
}
