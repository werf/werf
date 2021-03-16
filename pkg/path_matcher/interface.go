package path_matcher

type PathMatcher interface {
	// IsPathMatched checks for a complete matching of the path
	IsPathMatched(string) bool

	// ShouldGoThrough indicates that the directory or submodule path is not completely matched but may include matching files among the child files.
	// The method returns false if the path is completely matched.
	ShouldGoThrough(string) bool

	// IsDirOrSubmodulePathMatched returns true if IsPathMatched or ShouldGoThrough.
	// The method returns true if there is a possibility of containing the matching files among the child files.
	IsDirOrSubmodulePathMatched(string) bool

	// ID returns string that unambiguously defines the path matcher
	ID() string

	String() string
}
