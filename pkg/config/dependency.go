package config

import (
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/util"
)

type Dependency struct {
	ImageName string
	Before    string
	After     string
	Imports   []*DependencyImport

	raw *rawDependency
}

func (d *Dependency) validate() error {
	if d.ImageName == "" {
		return newDetailedConfigError("image name is not specified for dependency", d.raw, d.raw.doc())
	}

	switch imageType := d.raw.imageType(); imageType {
	case dependencyImageTypeStapel:
		switch {
		case d.Before == "" && d.After == "":
			return newDetailedConfigError("dependency stage is not specified with `before: install|setup` or `after: install|setup`", d.raw, d.raw.doc())
		case d.Before != "" && d.After != "":
			return newDetailedConfigError("only one dependency stage can be specified with `before: install|setup` or `after: install|setup`, but not both", d.raw, d.raw.doc())
		case d.Before != "" && d.Before != "install" && d.Before != "setup":
			return newDetailedConfigError(fmt.Sprintf("invalid dependency stage `before: %s`: expected install or setup!", d.Before), d.raw, d.raw.doc())
		case d.After != "" && d.After != "install" && d.After != "setup":
			return newDetailedConfigError(fmt.Sprintf("invalid dependency stage `after: %s`: expected install or setup!", d.After), d.raw, d.raw.doc())
		}
	case dependencyImageTypeDockerfile:
		switch {
		case d.Before != "":
			return newDetailedConfigError("`before:` directive is not supported for dockerfile dependencies", d.raw, d.raw.doc())
		case d.After != "":
			return newDetailedConfigError("`after:` directive is not supported for dockerfile dependencies", d.raw, d.raw.doc())
		}
	default:
		panic(fmt.Sprintf("unexpected imageType: %s", imageType))
	}

	var targetEnvs, targetBuildArgs []string
	for _, depImport := range d.Imports {
		if depImport.TargetEnv != "" {
			targetEnvs = append(targetEnvs, depImport.TargetEnv)
		}
		if depImport.TargetBuildArg != "" {
			targetBuildArgs = append(targetBuildArgs, depImport.TargetBuildArg)
		}
	}

	if len(targetEnvs) > 0 {
		duplicatedTargetEnvs := util.FindDuplicatedStrings(targetEnvs)
		if len(duplicatedTargetEnvs) > 0 {
			return newDetailedConfigError(fmt.Sprintf("each targetEnv for dependency import should be unique, but found duplicates for: %s", strings.Join(duplicatedTargetEnvs, ", ")), d.raw, d.raw.doc())
		}
	}

	if len(targetBuildArgs) > 0 {
		duplicatedTargetBuildArgs := util.FindDuplicatedStrings(targetBuildArgs)
		if len(duplicatedTargetBuildArgs) > 0 {
			return newDetailedConfigError(fmt.Sprintf("each targetBuildArg for dependency import should be unique, but found duplicates for: %s", strings.Join(duplicatedTargetBuildArgs, ", ")), d.raw, d.raw.doc())
		}
	}

	return nil
}
