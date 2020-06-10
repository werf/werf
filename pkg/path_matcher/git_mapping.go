package path_matcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/werf/pkg/util"

	"github.com/bmatcuk/doublestar"
)

func NewGitMappingPathMatcher(basePath string, includePaths, excludePaths []string, greedySearch bool) *GitMappingPathMatcher {
	return &GitMappingPathMatcher{
		basePathMatcher:  basePathMatcher{basePath: formatPath(basePath)},
		includePaths:     formatPaths(includePaths),
		excludePaths:     formatPaths(excludePaths),
		isGreedySearchOn: greedySearch,
	}
}

type GitMappingPathMatcher struct {
	basePathMatcher
	includePaths     []string
	excludePaths     []string
	isGreedySearchOn bool
}

func (f *GitMappingPathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *GitMappingPathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, includePaths=%v, excludePaths=%v, greedySearch=%v", f.basePath, f.includePaths, f.excludePaths, f.isGreedySearchOn)
}

func (f *GitMappingPathMatcher) MatchPath(path string) bool {
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

func (f *GitMappingPathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isMatched, shouldGoThrough := f.processDirOrSubmodulePath(formatPath(path))
	if f.isGreedySearchOn {
		return false, isMatched || shouldGoThrough
	} else {
		return isMatched, shouldGoThrough
	}
}

func (f *GitMappingPathMatcher) processDirOrSubmodulePath(path string) (bool, bool) {
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
	relPathParts := util.SplitFilepath(relPath)
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
	} else if strings.Contains(globParts[0], "**") {
		return glob, ""
	} else if len(globParts) > 1 {
		return filepath.Join(globParts[1:]...), ""
	} else {
		return "", glob
	}
}

func hasUniversalGlob(globs []string) bool {
	for _, glob := range globs {
		if glob == "." {
			return true
		}

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
