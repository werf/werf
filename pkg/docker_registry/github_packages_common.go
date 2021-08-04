package docker_registry

type gitHubPackagesBase struct {
	*defaultImplementation
	gitHubCredentials
}

type gitHubCredentials struct {
	token string
}

type gitHubPackagesOptions struct {
	defaultImplementationOptions
	gitHubCredentials
}

func newGitHubPackagesBase(options gitHubPackagesOptions) (*gitHubPackagesBase, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitHub := &gitHubPackagesBase{
		defaultImplementation: d,
		gitHubCredentials:     options.gitHubCredentials,
	}

	return gitHub, nil
}

func (r *gitHubPackagesBase) String() string {
	return GitHubPackagesImplementationName
}
