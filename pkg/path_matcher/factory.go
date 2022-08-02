package path_matcher

type PathMatcherOptions struct {
	BasePath             string
	IncludeGlobs         []string
	ExcludeGlobs         []string
	DockerignorePatterns []string
	Matchers             []PathMatcher
}

func NewPathMatcher(options PathMatcherOptions) PathMatcher {
	var matchers []PathMatcher

	if len(options.IncludeGlobs) != 0 {
		matchers = append(matchers, newIncludePathMatcher(options.IncludeGlobs))
	}

	if len(options.ExcludeGlobs) != 0 {
		matchers = append(matchers, newExcludePathMatcher(options.ExcludeGlobs))
	}

	if options.DockerignorePatterns != nil {
		matchers = append(matchers, newDockerfileIgnorePathMatcher(options.DockerignorePatterns))
	}

	if len(options.Matchers) != 0 {
		matchers = append(matchers, options.Matchers...)
	}

	if len(matchers) == 0 {
		matchers = append(matchers, NewTruePathMatcher())
	}

	var matcher PathMatcher
	matcher = NewMultiPathMatcher(matchers...)
	if options.BasePath != "" {
		matcher = newBasePathMatcher(options.BasePath, matcher)
	}

	return matcher
}
