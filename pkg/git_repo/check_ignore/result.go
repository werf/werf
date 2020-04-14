package check_ignore

import (
	"gopkg.in/src-d/go-git.v4"
)

type Result struct {
	repository            *git.Repository
	repositoryAbsFilepath string
	ignoredAbsFilepaths   []string
	submoduleResults      []*SubmoduleResult
}

type SubmoduleResult struct {
	*Result
}

func (r *Result) IgnoredFilesPaths() []string {
	paths := r.ignoredAbsFilepaths

	for _, submoduleResult := range r.submoduleResults {
		paths = append(paths, submoduleResult.ignoredAbsFilepaths...)
	}

	return paths
}
