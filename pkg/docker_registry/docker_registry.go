package docker_registry

import (
	"fmt"
	"regexp"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/flant/werf/pkg/image"
)

const (
	MultirepoRepoMode = "multirepo"
	MonorepoRepoMode  = "monorepo"
)

type DockerRegistry interface {
	CreateRepo(reference string) error
	DeleteRepo(reference string) error
	Tags(reference string) ([]string, error)
	GetRepoImage(reference string) (*image.Info, error)
	TryGetRepoImage(reference string) (*image.Info, error)
	IsRepoImageExists(reference string) (bool, error)
	GetRepoImageList(reference string) ([]*image.Info, error)
	SelectRepoImageList(reference string, f func(*image.Info) bool) ([]*image.Info, error)
	DeleteRepoImage(repoImageList ...*image.Info) error

	ResolveRepoMode(registryOrRepositoryAddress, repoMode string) (string, error)
	String() string
}

type DockerRegistryOptions struct {
	InsecureRegistry      bool
	SkipTlsVerifyRegistry bool
	DockerHubToken        string
	DockerHubUsername     string
	DockerHubPassword     string
	GitHubToken           string
	HarborUsername        string
	HarborPassword        string
}

func (o *DockerRegistryOptions) awsEcrOptions() awsEcrOptions {
	return awsEcrOptions{
		defaultImplementationOptions: o.defaultOptions(),
	}
}

func (o *DockerRegistryOptions) dockerHubOptions() dockerHubOptions {
	return dockerHubOptions{
		defaultImplementationOptions: o.defaultOptions(),
		dockerHubCredentials: dockerHubCredentials{
			token:    o.DockerHubToken,
			username: o.DockerHubUsername,
			password: o.DockerHubPassword,
		},
	}
}

func (o *DockerRegistryOptions) gcrOptions() GcrOptions {
	return GcrOptions{
		defaultImplementationOptions: o.defaultOptions(),
	}
}

func (o *DockerRegistryOptions) gitHubPackagesOptions() gitHubPackagesOptions {
	return gitHubPackagesOptions{
		defaultImplementationOptions: o.defaultOptions(),
		gitHubCredentials: gitHubCredentials{
			token: o.GitHubToken,
		},
	}
}

func (o *DockerRegistryOptions) gitLabRegistryOptions() gitLabRegistryOptions {
	return gitLabRegistryOptions{
		defaultImplementationOptions: o.defaultOptions(),
	}
}

func (o *DockerRegistryOptions) harborOptions() harborOptions {
	return harborOptions{
		defaultImplementationOptions: o.defaultOptions(),
		harborCredentials: harborCredentials{
			username: o.HarborUsername,
			password: o.HarborPassword,
		},
	}
}

func (o *DockerRegistryOptions) quayOptions() quayOptions {
	return quayOptions{
		defaultImplementationOptions: o.defaultOptions(),
	}
}

func (o *DockerRegistryOptions) defaultOptions() defaultImplementationOptions {
	return defaultImplementationOptions{apiOptions{
		InsecureRegistry:      o.InsecureRegistry,
		SkipTlsVerifyRegistry: o.SkipTlsVerifyRegistry,
	}}
}

func NewDockerRegistry(repositoryAddress string, implementation string, options DockerRegistryOptions) (DockerRegistry, error) {
	switch implementation {
	case AwsEcrImplementationName:
		return newAwsEcr(options.awsEcrOptions())
	case DockerHubImplementationName:
		return newDockerHub(options.dockerHubOptions())
	case GcrImplementationName:
		return newGcr(options.gcrOptions())
	case GitHubPackagesImplementationName:
		return newGitHubPackages(options.gitHubPackagesOptions())
	case GitLabRegistryImplementationName:
		return newGitLabRegistry(options.gitLabRegistryOptions())
	case HarborImplementationName:
		return newHarbor(options.harborOptions())
	case QuayImplementationName:
		return newQuay(options.quayOptions())
	case DefaultImplementationName:
		return newDefaultImplementation(options.defaultOptions())
	default:
		resolvedImplementation, err := ResolveImplementation(repositoryAddress, implementation)
		if err != nil {
			return nil, err
		}

		return NewDockerRegistry(repositoryAddress, resolvedImplementation, options)
	}
}

func ResolveImplementation(repository, implementation string) (string, error) {
	for _, supportedImplementation := range ImplementationList() {
		if supportedImplementation == implementation {
			return implementation, nil
		}
	}

	if implementation == "auto" || implementation == "" {
		return detectImplementation(repository)
	}

	return "", fmt.Errorf("docker registry implementation %s is not supported", implementation)
}

func detectImplementation(repositoryAddress string) (string, error) {
	parsedReference, err := name.NewTag(repositoryAddress)
	if err != nil {
		return "", err
	}

	for _, service := range []struct {
		name     string
		patterns []string
	}{
		{
			name:     AwsEcrImplementationName,
			patterns: awsEcrPatterns,
		},
		{
			name:     DockerHubImplementationName,
			patterns: dockerHubPatterns,
		},
		{
			name:     GcrImplementationName,
			patterns: gcrPatterns,
		},
		{
			name:     GitHubPackagesImplementationName,
			patterns: gitHubPackagesPatterns,
		},
		{
			name:     HarborImplementationName,
			patterns: harborPatterns,
		},
		{
			name:     QuayImplementationName,
			patterns: quayPatterns,
		},
	} {
		for _, pattern := range service.patterns {
			matched, err := regexp.MatchString(pattern, parsedReference.RegistryStr())
			if err != nil {
				return "", err
			}

			if matched {
				return service.name, nil
			}
		}
	}

	return "default", nil
}

func ImplementationList() []string {
	return []string{
		AwsEcrImplementationName,
		DefaultImplementationName,
		DockerHubImplementationName,
		GcrImplementationName,
		GitHubPackagesImplementationName,
		GitLabRegistryImplementationName,
		HarborImplementationName,
		QuayImplementationName,
	}
}
