package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/werf/werf/pkg/docker_registry"
)

type RepoData struct {
	IsCommon               bool
	DesignationStorageName string

	Implementation    *string // legacy
	ContainerRegistry *string
	DockerHubUsername *string
	DockerHubPassword *string
	DockerHubToken    *string
	GitHubToken       *string
	HarborUsername    *string
	HarborPassword    *string
	QuayToken         *string
}

func (d *RepoData) GetContainerRegistry() string {
	if *d.ContainerRegistry != "" {
		return *d.ContainerRegistry
	} else if *d.Implementation != "" {
		return *d.Implementation
	} else {
		return ""
	}
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

// legacy
func SetupImplementationForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
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

func SetupContainerRegistryForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usageTitle string
	if repoData.IsCommon {
		usageTitle = "Choose repo container registry"
	} else {
		usageTitle = fmt.Sprintf("Choose repo container registry for %s", repoData.DesignationStorageName)
	}

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

func SetupDockerHubUsernameForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Docker Hub username (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Docker Hub username for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.DockerHubUsername = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubUsername,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupDockerHubPasswordForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Docker Hub password (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Docker Hub password for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.DockerHubPassword = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubPassword,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupDockerHubTokenForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Docker Hub token (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Docker Hub token for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.DockerHubToken = new(string)
	cmd.Flags().StringVarP(
		repoData.DockerHubToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupGithubTokenForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("GitHub token (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("GitHub token for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.GitHubToken = new(string)
	cmd.Flags().StringVarP(
		repoData.GitHubToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupHarborUsernameForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Harbor username (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Harbor username for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.HarborUsername = new(string)
	cmd.Flags().StringVarP(
		repoData.HarborUsername,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupHarborPasswordForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Harbor password (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Harbor password for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.HarborPassword = new(string)
	cmd.Flags().StringVarP(
		repoData.HarborPassword,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)
}

func SetupQuayTokenForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("quay.io token (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("quay.io token for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

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
