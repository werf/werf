package path_matcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/fileutils"
)

func NewDockerfileIgnorePathMatcher(basePath string, patternMatcher *fileutils.PatternMatcher) *DockerfileIgnorePathMatcher {
	return &DockerfileIgnorePathMatcher{
		basePath:       basePath,
		patternMatcher: patternMatcher,
	}
}

type DockerfileIgnorePathMatcher struct {
	basePath       string
	patternMatcher *fileutils.PatternMatcher
}

func (f *DockerfileIgnorePathMatcher) BasePath() string {
	return f.basePath
}

func (f *DockerfileIgnorePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, patternMatcher=%v", f.basePath, f.patternMatcher.Patterns())
}

func (f *DockerfileIgnorePathMatcher) MatchPath(path string) bool {
	if !isRel(path, f.basePath) {
		return false
	} else if f.patternMatcher == nil || path == f.basePath {
		return true
	}

	relpath, err := filepath.Rel(f.basePath, path)
	if err != nil {
		panic(err)
	}

	isMatched, err := f.patternMatcher.Matches(relpath)
	if err != nil {
		panic(err)
	}

	return !isMatched
}

type pattern struct {
	pattern      string
	exclusion    bool
	isMatched    bool
	isInProgress bool
}

func (f *DockerfileIgnorePathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if f.patternMatcher == nil || len(f.patternMatcher.Patterns()) == 0 {
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
	var patterns []*pattern

	for _, p := range f.patternMatcher.Patterns() {
		patterns = append(patterns, &pattern{
			pattern:      p.String(),
			exclusion:    p.Exclusion(),
			isInProgress: true,
		})
	}

	for _, relPathPart := range relPathParts {
		for _, p := range patterns {
			if !p.isInProgress {
				continue
			}

			inProgressGlob, matchedGlob := matchGlob(relPathPart, p.pattern)
			if inProgressGlob != "" {
				p.pattern = inProgressGlob
			} else if matchedGlob != "" {
				p.isMatched = true
				p.isInProgress = false
			} else {
				p.isInProgress = false
			}
		}
	}

	isMatched := true
	shouldGoThrough := false
	for _, pattern := range patterns {
		if pattern.isMatched {
			isMatched = pattern.exclusion
			shouldGoThrough = false
		} else if pattern.isInProgress {
			isMatched = false
			shouldGoThrough = true
		}
	}

	return isMatched, shouldGoThrough
}

func (f *DockerfileIgnorePathMatcher) hasUniversalIgnoreGlob() bool {
	for _, pattern := range f.patternMatcher.Patterns() {
		if pattern.Exclusion() {
			continue
		}

		if hasUniversalGlob([]string{pattern.String()}) {
			return true
		}
	}

	return false
}
