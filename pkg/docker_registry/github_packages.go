package docker_registry

const GitHubPackagesImplementationName = "github"

var gitHubPackagesPatterns = []string{"^docker\\.pkg\\.github\\.com"}

type gitHubPackages struct {
	*defaultImplementation
}

type gitHubPackagesOptions struct {
	defaultImplementationOptions
}

func newGitHubPackages(options gitHubPackagesOptions) (*gitHubPackages, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackages{defaultImplementation: d}

	return gitHub, nil
}

func (r *gitHubPackagesOptions) String() string {
	return GitHubPackagesImplementationName
}
