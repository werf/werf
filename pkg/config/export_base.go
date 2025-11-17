package config

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/samber/lo"
)

type autoExcludeExport interface {
	AutoExcludeExportAndCheck(autoExcludeExport) bool
	GetIncludePathsForAutoExclude() []string
	GetExcludePathsForAutoExclude() []string
	AddExcludePath(string)

	GetTo() string
	GetRaw() interface{}
}

type ExportBase struct {
	Add          string
	To           string
	IncludePaths []string
	ExcludePaths []string
	Owner        string
	Group        string

	raw *rawExportBase
}

func (c *ExportBase) AutoExcludeExportAndCheck(exp autoExcludeExport) bool {
	if !isSubPath(c.GetTo(), exp.GetTo()) {
		return true
	}

	if len(c.GetIncludePathsForAutoExclude()) == 0 && len(exp.GetIncludePathsForAutoExclude()) == 0 {
		return false
	}

	for _, expIncludePath := range exp.GetIncludePathsForAutoExclude() {
		// If exact path is included in current export, do not exclude
		if slices.Contains(c.GetIncludePathsForAutoExclude(), expIncludePath) {
			return false
		}

		// If expIncludePath is a sub-path of any existing include, skip adding exclude
		if isSubPathOfSomePath(expIncludePath, c.GetIncludePathsForAutoExclude()) {
			continue
		}

		// If expIncludePath is covered by any existing exclude, skip
		if isSubPathOfSomePath(expIncludePath, c.GetExcludePathsForAutoExclude()) {
			continue
		}

		// Otherwise, calculate relative path and add to excludes
		extraExcludePath, err := filepath.Rel(path.Join(c.GetTo()), path.Join("/", expIncludePath)) // TODO rel
		if err != nil {
			panic(err)
		}

		c.AddExcludePath(extraExcludePath)
	}

	return true
}

// isSubPath checks if the given subPath is a sub-path of the given path
func isSubPath(subPath, path string) bool {
	subPathWithSlashEnding := strings.TrimRight(subPath, "/") + "/"
	return strings.HasPrefix(path, subPathWithSlashEnding) || path == subPath
}

// isSubPathOfSomePath checks if the given subPath is a sub-path of any path in the provided list
func isSubPathOfSomePath(subPath string, paths []string) bool {
	return lo.SomeBy(paths, func(p string) bool {
		return isSubPath(subPath, p)
	})
}

func (c *ExportBase) GetIncludePathsForAutoExclude() []string {
	var pathPrefix string
	if c.To != "/" {
		pathPrefix = c.To[1:len(c.To)]
	}

	if len(c.IncludePaths) == 0 && pathPrefix != "" {
		return []string{pathPrefix}
	} else {
		validateIncludePaths := make([]string, 0, len(c.IncludePaths))

		for _, p := range c.IncludePaths {
			validateIncludePaths = append(validateIncludePaths, path.Join(pathPrefix, p))
		}

		return validateIncludePaths
	}
}

func (c *ExportBase) GetExcludePathsForAutoExclude() []string {
	var pathPrefix string
	if c.To != "/" {
		pathPrefix = c.To[1:len(c.To)]
	}

	validateExcludePaths := make([]string, 0, len(c.ExcludePaths))

	for _, p := range c.ExcludePaths {
		validateExcludePaths = append(validateExcludePaths, path.Join(pathPrefix, p))
	}

	return validateExcludePaths
}

func (c *ExportBase) GetTo() string {
	return c.To
}

func (c *ExportBase) AddExcludePath(arg string) {
	c.ExcludePaths = append(c.ExcludePaths, arg)
}

func (c *ExportBase) GetRaw() interface{} {
	panic("should be implemented!")
}

func (c *ExportBase) validate() error {
	switch {
	case c.Add == "" || !isAbsolutePath(c.Add):
		return newDetailedConfigError("`add: PATH` absolute path required for import!", c.raw.rawOrigin.configSection(), c.raw.rawOrigin.doc())
	case c.To == "" || !isAbsolutePath(c.To):
		return newDetailedConfigError("`to: PATH` absolute path required for import!", c.raw.rawOrigin.configSection(), c.raw.rawOrigin.doc())
	case !allRelativePaths(c.IncludePaths):
		return newDetailedConfigError("`includePaths: [PATH, ...]|PATH` should be relative paths!", c.raw.rawOrigin.configSection(), c.raw.rawOrigin.doc())
	case !allRelativePaths(c.ExcludePaths):
		return newDetailedConfigError("`excludePaths: [PATH, ...]|PATH` should be relative paths!", c.raw.rawOrigin.configSection(), c.raw.rawOrigin.doc())
	default:
		return nil
	}
}
