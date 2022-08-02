package path_matcher

import (
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/util"
)

func newBasePathMatcher(basePath string, matcher PathMatcher) basePathMatcher {
	return basePathMatcher{basePath: formatPath(basePath), matcher: matcher}
}

type basePathMatcher struct {
	basePath string
	matcher  PathMatcher
}

func (m basePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)

	basePathCheck := util.IsSubpathOfBasePath(m.basePath, path) || m.basePath == path
	if !basePathCheck {
		return false
	}

	if m.matcher == nil {
		return true
	}

	return m.matcher.IsPathMatched(util.GetRelativeToBaseFilepath(m.basePath, path))
}

func (m basePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m basePathMatcher) ShouldGoThrough(path string) bool {
	path = formatPath(path)

	if util.IsSubpathOfBasePath(path, m.basePath) {
		return true
	}

	if m.matcher == nil {
		return false
	}

	return m.matcher.ShouldGoThrough(util.GetRelativeToBaseFilepath(m.basePath, path))
}

func (m basePathMatcher) ID() string {
	if m.basePath == "" && (m.matcher == nil || m.matcher.ID() == "") {
		return ""
	}

	var args []string
	args = append(args, "basePath")
	args = append(args, m.basePath)
	if m.matcher != nil {
		args = append(args, m.matcher.ID())
	}

	return util.Sha256Hash(strings.Join(args, ":::"))
}

func (m basePathMatcher) String() string {
	return fmt.Sprintf("{ basePath=%q, matcher=%q }", m.basePath, m.matcher.String())
}
