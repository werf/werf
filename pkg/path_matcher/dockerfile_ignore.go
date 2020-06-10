package path_matcher

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/pkg/fileutils"

	"github.com/werf/werf/pkg/util"
)

func NewDockerfileIgnorePathMatcher(basePath string, patternMatcher *fileutils.PatternMatcher, greedySearch bool) *DockerfileIgnorePathMatcher {
	return &DockerfileIgnorePathMatcher{
		basePathMatcher:  basePathMatcher{basePath: formatPath(basePath)},
		patternMatcher:   patternMatcher,
		isGreedySearchOn: greedySearch,
	}
}

type DockerfileIgnorePathMatcher struct {
	basePathMatcher
	patternMatcher   *fileutils.PatternMatcher
	isGreedySearchOn bool
}

func (f *DockerfileIgnorePathMatcher) BaseFilepath() string {
	return f.basePath
}

func (f *DockerfileIgnorePathMatcher) String() string {
	return fmt.Sprintf("basePath=`%s`, patternMatcher=%v, greedySearch=%v", f.basePath, f.patternMatcher.Patterns(), f.isGreedySearchOn)
}

func (f *DockerfileIgnorePathMatcher) MatchPath(path string) bool {
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

func (f *DockerfileIgnorePathMatcher) ProcessDirOrSubmodulePath(path string) (bool, bool) {
	isMatched, shouldGoThrough := f.processDirOrSubmodulePath(formatPath(path))
	if f.isGreedySearchOn {
		return false, isMatched || shouldGoThrough
	} else {
		return isMatched, shouldGoThrough
	}
}

func (f *DockerfileIgnorePathMatcher) processDirOrSubmodulePath(path string) (bool, bool) {
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
