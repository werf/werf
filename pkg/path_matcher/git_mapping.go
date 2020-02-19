package path_matcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
)

func NewGitMappingPathMatcher(basePath string, includePaths, excludePaths []string) *GitMappingPathMatcher {
	return &GitMappingPathMatcher{
		basePath:     basePath,
		includePaths: includePaths,
		excludePaths: excludePaths,
	}
}

type GitMappingPathMatcher struct {
	basePath     string
	includePaths []string
	excludePaths []string
}

func (f *GitMappingPathMatcher) BasePath() string {
	return f.basePath
}

func (f *GitMappingPathMatcher) IncludePaths() []string {
	return f.includePaths
}

func (f *GitMappingPathMatcher) ExcludePaths() []string {
	return f.excludePaths
}

func (f *GitMappingPathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, includePaths=%v, excludePaths=%v", f.basePath, f.includePaths, f.excludePaths)
}

func (f *GitMappingPathMatcher) MatchPath(path string) bool {
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

func (f *GitMappingPathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if len(f.includePaths) == 0 && len(f.excludePaths) == 0 {
			return true, false
		} else if hasUniversalGlob(f.excludePaths) {
			return false, false
		} else if hasUniversalGlob(f.includePaths) {
			return true, false
		} else if path == f.basePath {
			return false, true
		}
	} else if isBasePathRelativeToPath {
		return false, true
	} else { // path is not relative to basePath
		return false, false
	}

	relPath := rel(path, f.basePath)
	relPathParts := strings.Split(relPath, string(os.PathSeparator))
	inProgressIncludePaths := f.includePaths[:]
	inProgressExcludePaths := f.excludePaths[:]
	var matchedIncludePaths, matchedExcludePaths []string

	for _, pathPart := range relPathParts {
		if len(inProgressIncludePaths) == 0 && len(inProgressExcludePaths) == 0 {
			break
		}

		if len(inProgressIncludePaths) != 0 {
			inProgressIncludePaths, matchedIncludePaths = matchGlobs(pathPart, inProgressIncludePaths)
			if len(inProgressIncludePaths) == 0 && len(matchedIncludePaths) == 0 {
				return false, false
			}
		}

		if len(inProgressExcludePaths) != 0 {
			inProgressExcludePaths, matchedExcludePaths = matchGlobs(pathPart, inProgressExcludePaths)
			if len(inProgressExcludePaths) == 0 && len(matchedExcludePaths) == 0 {
				if len(inProgressIncludePaths) == 0 {
					return true, false
				}
			}
		}
	}

	if len(inProgressExcludePaths) != 0 {
		return false, !hasUniversalGlob(inProgressExcludePaths)
	} else if len(inProgressIncludePaths) != 0 {
		if hasUniversalGlob(inProgressIncludePaths) {
			return true, false
		} else {
			return false, true
		}
	} else if len(matchedExcludePaths) != 0 {
		return false, false
	} else if len(matchedIncludePaths) != 0 {
		return true, false
	} else {
		return false, false
	}
}

func (f *GitMappingPathMatcher) TrimFileBasePath(filePath string) string {
	return trimFileBasePath(filePath, f.basePath)
}

func matchGlobs(pathPart string, globs []string) (inProgressGlobs []string, matchedGlobs []string) {
	for _, glob := range globs {
		globParts := strings.Split(glob, string(os.PathSeparator))
		isMatched, err := doublestar.PathMatch(globParts[0], pathPart)
		if err != nil {
			panic(err)
		}

		if !isMatched {
			continue
		} else if strings.Contains(globParts[0], "**") {
			inProgressGlobs = append(inProgressGlobs, glob)
		} else if len(globParts) > 1 {
			inProgressGlobs = append(inProgressGlobs, filepath.Join(globParts[1:]...))
		} else {
			matchedGlobs = append(matchedGlobs, glob)
		}
	}

	return
}

func hasUniversalGlob(globs []string) bool {
	for _, glob := range globs {
		if strings.TrimRight(glob, "*"+string(os.PathSeparator)) == "" {
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

func rel(targetPath, basePath string) string {
	if basePath == "" {
		return targetPath
	}

	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		panic(err)
	}

	return relPath
}

func trimFileBasePath(filePath, basePath string) string {
	if filePath == basePath {
		// filePath path is always a path to a file, not a directory.
		// Thus if basePath is equal filePath, then basePath is a path to the file.
		// Return file name in this case by convention.
		return filepath.Base(filePath)
	}

	return rel(filePath, basePath)
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

func isPathMatched(filePath, pattern string) bool {
	matched, err := doublestar.PathMatch(pattern, filePath)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	matched, err = doublestar.PathMatch(filepath.Join(pattern, "**", "*"), filePath)
	if err != nil {
		panic(err)
	}
	if matched {
		return true
	}

	return false
}
