package config

import (
	"path"
	"path/filepath"
	"strings"
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
	if !isSubPath(c.To, exp.GetTo()) {
		return true
	}

	if len(c.GetIncludePathsForAutoExclude()) == 0 && len(exp.GetIncludePathsForAutoExclude()) == 0 {
		return false
	}

loop:
	for _, expIncludePath := range exp.GetIncludePathsForAutoExclude() {
		for _, includePath := range c.GetIncludePathsForAutoExclude() {
			if expIncludePath == includePath {
				return false
			}

			if isSubPath(expIncludePath, includePath) {
				continue loop
			}
		}

		for _, excludePath := range c.GetExcludePathsForAutoExclude() {
			if isSubPath(excludePath, expIncludePath) {
				continue loop
			}
		}

		extraExcludePath, err := filepath.Rel(path.Join(c.GetTo()), path.Join("/", expIncludePath)) // TODO rel
		if err != nil {
			panic(err)
		}

		c.AddExcludePath(extraExcludePath)
	}

	return true
}

func isSubPath(subPath, path string) bool {
	subPathWithSlashEnding := strings.TrimRight(subPath, "/") + "/"
	return strings.HasPrefix(path, subPathWithSlashEnding) || path == subPath
}

func (c *ExportBase) GetIncludePathsForAutoExclude() []string {
	var pathPrefix string
	if c.To != "/" {
		pathPrefix = c.To[1:len(c.To)]
	}

	if len(c.IncludePaths) == 0 && pathPrefix != "" {
		return []string{pathPrefix}
	} else {
		var validateIncludePaths []string
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

	var validateExcludePaths []string
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
