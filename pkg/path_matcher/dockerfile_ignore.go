package path_matcher

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/moby/patternmatcher"

	"github.com/werf/common-go/pkg/util"
)

func NewDockerfileIgnorePathMatcher(ctx context.Context, dockerignorePatterns []string) (PathMatcher, error) {
	for _, pattern := range dockerignorePatterns {
		m, err := patternmatcher.New([]string{pattern})
		if err != nil {
			return nil, fmt.Errorf("parse ignore pattern %q: %w", pattern, err)
		}
		// Moby compiles patterns lazily; force compilation of exclusion patterns too.
		if _, err := m.MatchesUsingParentResult(".", m.Exclusions()); err != nil {
			return nil, fmt.Errorf("compile ignore pattern %q: %w", pattern, err)
		}
	}

	m, err := patternmatcher.New(dockerignorePatterns)
	if err != nil {
		return nil, fmt.Errorf("create ignore matcher: %w", err)
	}

	return dockerfileIgnorePathMatcher{
		patternMatcher: m,
	}, nil
}

var _ PathMatcher = dockerfileIgnorePathMatcher{}

type dockerfileIgnorePathMatcher struct {
	patternMatcher *patternmatcher.PatternMatcher
}

func (m dockerfileIgnorePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)
	if m.patternMatcher == nil {
		return true
	}

	isMatched, err := m.patternMatcher.Matches(path)
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

func (m dockerfileIgnorePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m dockerfileIgnorePathMatcher) ShouldGoThrough(path string) bool {
	return m.shouldGoThrough(formatPath(path))
}

func (m dockerfileIgnorePathMatcher) shouldGoThrough(path string) bool {
	if m.patternMatcher == nil || len(m.patternMatcher.Patterns()) == 0 {
		return false
	}

	if path == "" {
		return true
	}

	pathParts := util.SplitFilepath(path)
	var patterns []*pattern
	for _, p := range m.patternMatcher.Patterns() {
		patterns = append(patterns, &pattern{
			pattern:      p.String(),
			exclusion:    p.Exclusion(),
			isInProgress: true,
		})
	}

	for _, pathPart := range pathParts {
		for _, p := range patterns {
			if !p.isInProgress {
				continue
			}

			inProgressGlob, matchedGlob := matchGlob(pathPart, p.pattern)
			switch {
			case inProgressGlob != "":
				p.pattern = inProgressGlob
			case matchedGlob != "":
				p.isMatched = true
				p.isInProgress = false
			default:
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

func (m dockerfileIgnorePathMatcher) ID() string {
	var cleanedPatterns []string
	for _, pattern := range m.patternMatcher.Patterns() {
		cleanedPatterns = append(cleanedPatterns, pattern.String())
	}

	if len(cleanedPatterns) == 0 {
		return ""
	}

	sort.Strings(cleanedPatterns)

	var args []string
	args = append(args, "dockerfileIgnore")
	args = append(args, cleanedPatterns...)
	return util.Sha256Hash(strings.Join(args, ":::"))
}

func (m dockerfileIgnorePathMatcher) String() string {
	return fmt.Sprintf("{ patternMatcher=%v }", m.patternMatcher.Patterns())
}
