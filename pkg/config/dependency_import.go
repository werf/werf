package config

import (
	"fmt"
	"strings"
)

type DependencyImport struct {
	Type           DependencyImportType
	TargetBuildArg string
	TargetEnv      string

	// TODO: raw *rawDependencyImport
}

func (i *DependencyImport) validate(img ImageInterface) error {
	switch {
	case img.IsStapel() && i.TargetBuildArg != "":
		return newDetailedConfigError("`targetBuildArg cannot be used in the stapel image", nil, nil) // TODO: raw
	case !img.IsStapel() && i.TargetEnv != "":
		return newDetailedConfigError("`targetEnv cannot be used in the dockerfile image", nil, nil) // TODO: raw
	}

	switch i.Type {
	case ImageNameImport, ImageTagImport, ImageRepoImport:
	default:
		return newDetailedConfigError(fmt.Sprintf("invalid `type: %s` for dependency import, expected one of: %s", i.Type, strings.Join([]string{string(ImageNameImport), string(ImageTagImport), string(ImageRepoImport)}, ", ")), nil, nil) // TODO: raw
	}

	return nil
}
