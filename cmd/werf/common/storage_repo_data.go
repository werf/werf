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

	Implementation    *string
	DockerHubUsername *string
	DockerHubPassword *string
	DockerHubToken    *string
	GitHubToken       *string
	HarborUsername    *string
	HarborPassword    *string
	QuayToken         *string
}

func MergeRepoData(repoDataArr ...*RepoData) *RepoData {
	res := &RepoData{}

	for _, repoData := range repoDataArr {
		if res.Implementation == nil || *res.Implementation == "" {
			res.Implementation = repoData.Implementation
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

func SetupImplementationForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usageTitle string
	if repoData.IsCommon {
		usageTitle = "Choose common repo implementation for any stages storage or images repo specified for the command"
	} else {
		usageTitle = fmt.Sprintf("Choose repo implementation for %s", repoData.DesignationStorageName)
	}

	repoData.Implementation = new(string)
	cmd.Flags().StringVarP(
		repoData.Implementation,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		fmt.Sprintf(`%s.
The following docker registry implementations are supported: %s.
Default %s or auto mode (detect implementation by a registry).`,
			usageTitle,
			strings.Join(docker_registry.ImplementationList(), ", "),
			strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "),
		),
	)
}

func SetupDockerHubUsernameForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Common Docker Hub username for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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
		usage = fmt.Sprintf("Common Docker Hub password for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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
		usage = fmt.Sprintf("Common Docker Hub token for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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
		usage = fmt.Sprintf("Common GitHub token for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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
		usage = fmt.Sprintf("Common Harbor username for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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

	_ = cmd.Flags().MarkHidden(paramName)
}

func SetupHarborPasswordForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Common Harbor password for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
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

	_ = cmd.Flags().MarkHidden(paramName)
}

func SetupQuayTokenForRepoData(repoData *RepoData, cmd *cobra.Command, paramName string, paramEnvNames []string) {
	var usage string
	if repoData.IsCommon {
		usage = fmt.Sprintf("Common quay.io token for any stages storage or images repo specified for the command (default %s)", strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	} else {
		usage = fmt.Sprintf("Common quay.io token for %s (default %s)", repoData.DesignationStorageName, strings.Join(getParamEnvNamesForUsageDescription(paramEnvNames), ", "))
	}

	repoData.QuayToken = new(string)
	cmd.Flags().StringVarP(
		repoData.QuayToken,
		paramName,
		"",
		getDefaultValueByParamEnvNames(paramEnvNames),
		usage,
	)

	_ = cmd.Flags().MarkHidden(paramName)
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
