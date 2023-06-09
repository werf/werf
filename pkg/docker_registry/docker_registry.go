package docker_registry

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

type DockerRegistryOptions struct {
	InsecureRegistry      bool
	SkipTlsVerifyRegistry bool
	DockerHubToken        string
	DockerHubUsername     string
	DockerHubPassword     string
	GitHubToken           string
	HarborUsername        string
	HarborPassword        string
	QuayToken             string
	SelectelAccount       string
	SelectelVPC           string
	SelectelVPCID         string
	SelectelUsername      string
	SelectelPassword      string
}

func (o *DockerRegistryOptions) awsEcrOptions() awsEcrOptions {
	return awsEcrOptions{
		defaultImplementationOptions: o.defaultOptions(),
	}
}

func (o *DockerRegistryOptions) azureAcrOptions() azureCrOptions {
	return azureCrOptions{
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
		quayCredentials: quayCredentials{
			token: o.QuayToken,
		},
	}
}

func (o *DockerRegistryOptions) selectelOptions() selectelOptions {
	return selectelOptions{
		defaultImplementationOptions: o.defaultOptions(),
		selectelCredentials: selectelCredentials{
			username: o.SelectelUsername,
			password: o.SelectelPassword,
			account:  o.SelectelAccount,
			vpc:      o.SelectelVPC,
			vpcID:    o.SelectelVPCID,
		},
	}
}

func (o *DockerRegistryOptions) defaultOptions() defaultImplementationOptions {
	return defaultImplementationOptions{apiOptions{
		InsecureRegistry:      o.InsecureRegistry,
		SkipTlsVerifyRegistry: o.SkipTlsVerifyRegistry,
	}}
}

func NewDockerRegistry(repositoryAddress, implementation string, options DockerRegistryOptions) (Interface, error) {
	var res Interface
	var err error

	res, err = newDockerRegistry(repositoryAddress, implementation, options)
	if err != nil {
		return nil, err
	}
	if debugDockerRegistry() {
		res = NewDockerRegistryTracer(res, nil)
	}

	res = newDockerRegistryWithCache(res)
	return res, nil
}

func newDockerRegistry(repositoryAddress, implementation string, options DockerRegistryOptions) (Interface, error) {
	if err := ValidateRepositoryReference(repositoryAddress); err != nil {
		return nil, err
	}

	switch implementation {
	case AwsEcrImplementationName:
		return newAwsEcr(options.awsEcrOptions())
	case AzureCrImplementationName:
		return newAzureCr(options.azureAcrOptions())
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
	case SelectelImplementationName:
		selectelCR, errCR := newSelectel(options.selectelOptions())
		_, _, repository, err := selectelCR.parseReference(repositoryAddress)
		if err != nil {
			return nil, err
		}
		if repository == "" {
			return nil, fmt.Errorf("%s implementation is buggy. Add repository to WERF_REPO variable. Example: %s/project", implementation, repositoryAddress)
		}
		return selectelCR, errCR
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

func detectImplementation(accountOrRepositoryAddress string) (string, error) {
	var parsedResource authn.Resource
	var err error

	parts := strings.SplitN(accountOrRepositoryAddress, "/", 2)
	if len(parts) == 1 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		parsedResource, err = name.NewRegistry(accountOrRepositoryAddress)
		if err != nil {
			return "", err
		}
	} else {
		parsedResource, err = name.NewRepository(accountOrRepositoryAddress)
		if err != nil {
			return "", err
		}
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
			name:     AzureCrImplementationName,
			patterns: azureCrPatterns,
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
			name:     GitLabRegistryImplementationName,
			patterns: gitlabPatterns,
		},
		{
			name:     HarborImplementationName,
			patterns: harborPatterns,
		},
		{
			name:     QuayImplementationName,
			patterns: quayPatterns,
		},
		{
			name:     SelectelImplementationName,
			patterns: selectelPatterns,
		},
	} {
		for _, pattern := range service.patterns {
			matched, err := regexp.MatchString(pattern, parsedResource.RegistryStr())
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
		AzureCrImplementationName,
		DefaultImplementationName,
		DockerHubImplementationName,
		GcrImplementationName,
		GitHubPackagesImplementationName,
		GitLabRegistryImplementationName,
		HarborImplementationName,
		QuayImplementationName,
		SelectelImplementationName,
	}
}
