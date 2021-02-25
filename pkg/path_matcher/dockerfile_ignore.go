package path_matcher

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/pkg/fileutils"

	"github.com/werf/werf/pkg/util"
)

func NewDockerfileIgnorePathMatcher(basePath string, patternMatcher *fileutils.PatternMatcher) *DockerfileIgnorePathMatcher {
	return &DockerfileIgnorePathMatcher{
		basePathMatcher: basePathMatcher{basePath: formatPath(basePath)},
		patternMatcher:  patternMatcher,
	}
}

type DockerfileIgnorePathMatcher struct {
	basePathMatcher
	patternMatcher *fileutils.PatternMatcher
}

func (f *DockerfileIgnorePathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *DockerfileIgnorePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, patternMatcher=%v", f.basePath, f.patternMatcher.Patterns())
}

func (f *DockerfileIgnorePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)

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

func (f *DockerfileIgnorePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return f.IsPathMatched(path) || f.ShouldGoThrough(path)
}

func (f *DockerfileIgnorePathMatcher) ShouldGoThrough(path string) bool {
	return f.shouldGoThrough(formatPath(path))
}

func (f *DockerfileIgnorePathMatcher) shouldGoThrough(path string) bool {
	isBasePathRelativeToPath := isSubDirOf(f.basePath, path)
	isPathRelativeToBasePath := isSubDirOf(path, f.basePath)

	if isPathRelativeToBasePath || path == f.basePath {
		if f.patternMatcher == nil || len(f.patternMatcher.Patterns()) == 0 {
			return false
		} else if path == f.basePath {
			return true
		}

		return f.shouldGoThroughDetailedCheck(path)
	} else if isBasePathRelativeToPath {
		return true
	} else { // path is not relative to basePath
		return false
	}
}

func (f *DockerfileIgnorePathMatcher) shouldGoThroughDetailedCheck(path string) bool {
	relPath := rel(path, f.basePath)
	relPathParts := util.SplitFilepath(relPath)
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

	shouldGoThrough := false
	for _, pattern := range patterns {
		if pattern.isMatched {
			shouldGoThrough = false
		} else if pattern.isInProgress {
			shouldGoThrough = true
		}
	}

	return shouldGoThrough
}
