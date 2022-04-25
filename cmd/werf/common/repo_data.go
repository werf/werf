package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
)

func (repoData *RepoData) CreateDockerRegistry(insecureRegistry, skipTlsVerifyRegistry bool) (docker_registry.Interface, error) {
	cr := repoData.GetContainerRegistry()
	if err := ValidateRepoContainerRegistry(cr); err != nil {
		return nil, err
	}

	regOpts := repoData.GetDockerRegistryOptions(insecureRegistry, skipTlsVerifyRegistry)
	dockerRegistry, err := docker_registry.NewDockerRegistry(*repoData.Address, cr, regOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating container registry accessor for repo %q: %w", *repoData.Address, err)
	}

	return dockerRegistry, nil
}

func (repoData *RepoData) CreateStagesStorage(containerBackend container_backend.ContainerBackend, insecureRegistry, skipTlsVerifyRegistry bool) (storage.StagesStorage, error) {
	if *repoData.Address == storage.LocalStorageAddress {
		return storage.NewDockerServerStagesStorage(containerBackend.(*container_backend.DockerServerBackend)), nil
	} else {
		dockerRegistry, err := repoData.CreateDockerRegistry(insecureRegistry, skipTlsVerifyRegistry)
		if err != nil {
			return nil, fmt.Errorf("error creating container registry accessor for repo %s: %w", *repoData.Address, err)
		}
		return storage.NewRepoStagesStorage(*repoData.Address, containerBackend, dockerRegistry), nil
	}
}

type RepoData struct {
	Name string

	Address           *string
	Implementation    *string // legacy
	ContainerRegistry *string
	DockerHubUsername *string
	DockerHubPassword *string
	DockerHubToken    *string
	GitHubToken       *string
	HarborUsername    *string
	HarborPassword    *string
	QuayToken         *string

	RepoDataOptions
}

type RepoDataOptions struct {
	OnlyAddress bool
}

func NewRepoData(name string, opts RepoDataOptions) *RepoData {
	return &RepoData{Name: name, RepoDataOptions: opts}
}

func (d *RepoData) GetContainerRegistry() string {
	if d.OnlyAddress {
		return ""
	}

	switch {
	case *d.ContainerRegistry != "":
		return *d.ContainerRegistry
	case *d.Implementation != "":
		logboek.Context(GetContext()).Warn().LogLn("DEPRECATION WARNING: The option --repo-implementation ($WERF_REPO_IMPLEMENTATION) is renamed to --repo-container-registry ($WERF_REPO_CONTAINER_REGISTRY) and will be removed in v1.3!")
		return *d.Implementation
	default:
		return ""
	}
}

func (d *RepoData) GetDockerRegistryOptions(insecureRegistry, skipTlsVerifyRegistry bool) docker_registry.DockerRegistryOptions {
	opts := docker_registry.DockerRegistryOptions{
		InsecureRegistry:      insecureRegistry,
		SkipTlsVerifyRegistry: skipTlsVerifyRegistry,
	}

	if !d.OnlyAddress {
		opts.DockerHubUsername = *d.DockerHubUsername
		opts.DockerHubPassword = *d.DockerHubPassword
		opts.DockerHubToken = *d.DockerHubToken
		opts.GitHubToken = *d.GitHubToken
		opts.HarborUsername = *d.HarborUsername
		opts.HarborPassword = *d.HarborPassword
		opts.QuayToken = *d.QuayToken
	}

	return opts

	return docker_registry.DockerRegistryOptions{
		DockerHubUsername: *d.DockerHubUsername,
		DockerHubPassword: *d.DockerHubPassword,
		DockerHubToken:    *d.DockerHubToken,
		GitHubToken:       *d.GitHubToken,
		HarborUsername:    *d.HarborUsername,
		HarborPassword:    *d.HarborPassword,
		QuayToken:         *d.QuayToken,
	}
}

func (repoData *RepoData) SetupCmd(cmd *cobra.Command) {
	repoNameUpper := strings.ToUpper(strings.ReplaceAll(repoData.Name, "-", "_"))
	makeEnvVar := func(variableSuffix string) string {
		return fmt.Sprintf("WERF_%s_%s", repoNameUpper, variableSuffix)
	}
	makeOpt := func(optSuffix string) string {
		return fmt.Sprintf("%s-%s", repoData.Name, optSuffix)
	}

	repoData.SetupAddressForRepoData(cmd, repoData.Name, []string{fmt.Sprintf("WERF_%s", repoNameUpper)})
	if repoData.OnlyAddress {
		return
	}

	repoData.SetupImplementationForRepoData(cmd, makeOpt("implementation"), []string{makeEnvVar("IMPLEMENTATION")}) // legacy
	repoData.SetupContainerRegistryForRepoData(cmd, makeOpt("container-registry"), []string{makeEnvVar("CONTAINER_REGISTRY")})
	repoData.SetupDockerHubUsernameForRepoData(cmd, makeOpt("docker-hub-username"), []string{makeEnvVar("DOCKER_HUB_USERNAME")})
	repoData.SetupDockerHubPasswordForRepoData(cmd, makeOpt("docker-hub-password"), []string{makeEnvVar("DOCKER_HUB_PASSWORD")})
	repoData.SetupDockerHubTokenForRepoData(cmd, makeOpt("docker-hub-token"), []string{makeEnvVar("DOCKER_HUB_TOKEN")})
	repoData.SetupGithubTokenForRepoData(cmd, makeOpt("github-token"), []string{makeEnvVar("GITHUB_TOKEN")})
	repoData.SetupHarborUsernameForRepoData(cmd, makeOpt("harbor-username"), []string{makeEnvVar("HARBOR_USERNAME")})
	repoData.SetupHarborPasswordForRepoData(cmd, makeOpt("harbor-password"), []string{makeEnvVar("HARBOR_PASSWORD")})
	repoData.SetupQuayTokenForRepoData(cmd, makeOpt("quay-token"), []string{makeEnvVar("QUAY_TOKEN")})
}

func MergeRepoData(repoDataArr ...*RepoData) *RepoData {
	res := &RepoData{}

	for _, repoData := range repoDataArr {
		if res.GetContainerRegistry() == "" {
			value := repoData.GetContainerRegistry()
			res.ContainerRegistry = &value
		}
		if res.DockerHubUsername == nil || *res.DockerHubUsername == "" {
			res.DockerHubUsername = repoData.DockerHubUsername
		}
		if res.DockerHubPassword == nil || *res.DockerHubPassword == "" {
			res.DockerHubPassword = repoData.DockerHubPassword
		}
		if res.DockerHubToken == nil || *res.DockerHubToken == "" {
			res.DockerHubToken = repoData.DockerHubToken
		}
		if res.GitHubToken == nil || *res.GitHubToken == "" {
			res.GitHubToken = repoData.GitHubToken
		}
		if res.HarborUsername == nil || *res.HarborUsername == "" {
			res.HarborUsername = repoData.HarborUsername
		}
		if res.HarborPassword == nil || *res.HarborPassword == "" {
			res.HarborPassword = repoData.HarborPassword
		}
		if res.QuayToken == nil || *res.QuayToken == "" {
			res.QuayToken = repoData.QuayToken
		}
	}

	return res
}

func (repoData *RepoData) SetupAddressForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("Container registry storage address (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.Address = new(string)
	cmd.Flags().StringVarP(
		repoData.Address,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

// legacy
func (repoData *RepoData) SetupImplementationForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	repoData.Implementation = new(string)
	cmd.Flags().StringVarP(
		repoData.Implementation,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		"",
	)
	cmd.Flag(paramName).Hidden = true
}

func (repoData *RepoData) SetupContainerRegistryForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usageTitle := fmt.Sprintf("Choose %s container registry implementation", repoData.Name)

	repoData.ContainerRegistry = new(string)
	cmd.Flags().StringVarP(
		repoData.ContainerRegistry,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		fmt.Sprintf(`%s.
The following container registries are supported: %s.
Default %s or auto mode (detect container registry by repo address).`,
			usageTitle,
			strings.Join(docker_registry.ImplementationList(), ", "),
			strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "),
		),
	)
}

func (repoData *RepoData) SetupDockerHubUsernameForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s Docker Hub username (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.DockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubUsername,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupDockerHubPasswordForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s Docker Hub password (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.DockerHubPassword = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubPassword,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupDockerHubTokenForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s Docker Hub token (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.DockerHubToken = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupGithubTokenForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s GitHub token (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.GitHubToken = new(string)
	cmd.Flags().StringVarP(
		repoData.GitHubToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupHarborUsernameForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s Harbor username (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.HarborUsername = new(string)
	cmd.Flags().StringVarP(
		repoData.HarborUsername,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupHarborPasswordForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s Harbor password (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.HarborPassword = new(string)
	cmd.Flags().StringVarP(
		repoData.HarborPassword,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func (repoData *RepoData) SetupQuayTokenForRepoData(cmd *cobra.Command, paramName string, paramEnvNames []string) {
	usage := fmt.Sprintf("%s quay.io token (default %s)", repoData.Name, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))

	repoData.QuayToken = new(string)
	cmd.Flags().StringVarP(
		repoData.QuayToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func getDefaultValueByParamEnvNames(paramEnvNames []string) string {
	var defaultValue string
	for _, paramEnvName := range paramEnvNames {
		value := os.Getenv(paramEnvName)
		if value != "" {
			defaultValue = value
			break
		}
	}
	return defaultValue
}

func getParamEnvNamesForUsageDescription(paramEnvNames []string) []string {
	paramEnvNamesWithDollar := []string{}
	for _, paramEnvName := range paramEnvNames {
		paramEnvNamesWithDollar = append(paramEnvNamesWithDollar, "$"+paramEnvName)
	}
	return paramEnvNamesWithDollar
}

func ValidateRepoContainerRegistry(containerRegistry string) error {
	supportedValues := docker_registry.ImplementationList()
	supportedValues = append(supportedValues, "auto", "")

	for _, supportedContainerRegistry := range supportedValues {
		if supportedContainerRegistry == containerRegistry {
			return nil
		}
	}

	return fmt.Errorf("specified container registry %q is not supported", containerRegistry)
}
