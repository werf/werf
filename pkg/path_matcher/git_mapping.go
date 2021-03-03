package path_matcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/util"

	"github.com/bmatcuk/doublestar"
)

func NewGitMappingPathMatcher(basePath string, includePaths, excludePaths []string) *GitMappingPathMatcher {
	return &GitMappingPathMatcher{
		basePathMatcher: basePathMatcher{basePath: formatPath(basePath)},
		includePaths:    formatPaths(includePaths),
		excludePaths:    formatPaths(excludePaths),
	}
}

type GitMappingPathMatcher struct {
	basePathMatcher
	includePaths []string
	excludePaths []string
}

func (f *GitMappingPathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *GitMappingPathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, includePaths=%v, excludePaths=%v", f.basePath, f.includePaths, f.excludePaths)
}

func (f *GitMappingPathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)

	if !isRel(path, f.basePath) {
		return false
	}

	subPathOrDot := rel(path, f.basePath)

	if len(f.includePaths) > 0 {
		if !isAnyPatternMatched(subPathOrDot, f.includePaths) {
			return false
		}
	}

	if len(f.excludePaths) > 0 {
		if isAnyPatternMatched(subPathOrDot, f.excludePaths) {
			return false
		}
	}

	return true
}

func (f *GitMappingPathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return f.IsPathMatched(path) || f.ShouldGoThrough(path)
}

func (f *GitMappingPathMatcher) ShouldGoThrough(path string) bool {
	return f.shouldGoThrough(formatPath(path))
}

func (f *GitMappingPathMatcher) shouldGoThrough(path string) bool {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if len(f.includePaths) == 0 && len(f.excludePaths) == 0 {
			return false
		}
		if hasUniversalGlob(f.excludePaths) {
			return false
		}
		if hasUniversalGlob(f.includePaths) {
			return false
		}
		if path == f.basePath {
			return true
		}

		return f.shouldGoThroughDetailedCheck(path)
	} else if isBasePathRelativeToPath {
		return true
	} else { // path is not relative to basePath
		return false
	}
}

func (f *GitMappingPathMatcher) shouldGoThroughDetailedCheck(path string) bool {
	relPath := rel(path, f.basePath)
	relPathParts := util.SplitFilepath(relPath)
	inProgressIncludePaths := f.includePaths
	inProgressExcludePaths := f.excludePaths
	var matchedIncludePaths, matchedExcludePaths []string

	for _, pathPart := range relPathParts {
		if len(inProgressIncludePaths) == 0 && len(inProgressExcludePaths) == 0 {
			break
		}

		if len(inProgressIncludePaths) != 0 {
			inProgressIncludePaths, matchedIncludePaths = matchGlobs(pathPart, inProgressIncludePaths)
			if len(inProgressIncludePaths) == 0 && len(matchedIncludePaths) == 0 {
				return false
			}
		}

		if len(inProgressExcludePaths) != 0 {
			inProgressExcludePaths, matchedExcludePaths = matchGlobs(pathPart, inProgressExcludePaths)
			if len(inProgressExcludePaths) == 0 && len(matchedExcludePaths) == 0 {
				if len(inProgressIncludePaths) == 0 {
					return false
				}
			}
		}
	}

	if len(inProgressExcludePaths) != 0 || len(inProgressIncludePaths) != 0 {
		return !hasUniversalGlob(inProgressExcludePaths)
	}
	if len(matchedExcludePaths) != 0 || len(matchedIncludePaths) != 0 {
		return false
	}

	return false
}

func matchGlobs(pathPart string, globs []string) (inProgressGlobs []string, matchedGlobs []string) {
	for _, glob := range globs {
		inProgressGlob, matchedGlob := matchGlob(pathPart, glob)
		if inProgressGlob != "" {
			inProgressGlobs = append(inProgressGlobs, inProgressGlob)
		} else if matchedGlob != "" {
			matchedGlobs = append(matchedGlobs, matchedGlob)
		}
	}

	return
}

func matchGlob(pathPart string, glob string) (inProgressGlob, matchedGlob string) {
	globParts := strings.Split(glob, string(os.PathSeparator))
	isMatched, err := doublestar.PathMatch(globParts[0], pathPart)
	if err != nil {
		panic(err)
	}
	if !isMatched {
		return "", ""
	}
	if strings.Contains(globParts[0], "**") {
		return glob, ""
	}
	if len(globParts) > 1 {
		return filepath.Join(globParts[1:]...), ""
	}
	return "", glob
}

func hasUniversalGlob(globs []string) bool {
	for _, glob := range globs {
		if glob == "." {
			return true
		}

		if trimRightAsterisks(glob) == "" {
			return true
		}
	}

	return false
}

func isSubDirOf(targetPath, basePath string) bool {
	if targetPath == basePath {
		return false
	}

	return isRel(targetPath, basePath)
}

func isRel(targetPath, basePath string) bool {
	if basePath == "" {
		return true
	}

	return strings.HasPrefix(targetPath+string(os.PathSeparator), basePath+string(os.PathSeparator))
}

func isAnyPatternMatched(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if isRel(path, pattern) {
			return true
		}

		if isPathMatched(path, pattern) {
			return true
		}
	}

	return false
}

func isPathMatched(filePath, glob string) bool {
	// The glob as-is.
	// The glob without asterisks on the right (path/*/dir/**/*, path/*/dir/**, path/*/dir/*/*/** -> path/*/dir).
	// The previous glob with the universal part `**/*` (path/*/dir/**/*).
	for _, g := range []string{
		glob,
		trimRightAsterisks(glob),
		filepath.Join(trimRightAsterisks(glob), "**", "*"),
	} {
		matched, err := doublestar.PathMatch(g, filePath)
		if err != nil {
			panic(err)
		}

		if matched {
			return true
		}
	}

	return false
}

func trimRightAsterisks(pattern string) string {
	return strings.TrimRight(pattern, "*\\/")
}
