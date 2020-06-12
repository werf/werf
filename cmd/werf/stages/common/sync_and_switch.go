package common

import (
	"fmt"
	"os"

	"github.com/werf/werf/cmd/werf/common"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/storage"
	"github.com/spf13/cobra"
)

type SyncCmdData struct {
	FromStagesStorage         *string
	FromStagesStorageRepoData *common.RepoData

	ToStagesStorage         *string
	ToStagesStorageRepoData *common.RepoData

	RemoveSource      *bool
	CleanupLocalCache *bool
}

func SetupRemoveSource(cmdData *SyncCmdData, cmd *cobra.Command) {
	cmdData.RemoveSource = new(bool)
	cmd.Flags().BoolVarP(cmdData.RemoveSource, "remove-source", "", common.GetBoolEnvironmentDefaultFalse("WERF_REMOVE_SOURCE"), "Remove existing project stages from source stages storage during sync procedure (default $WERF_REMOVE_SOURCE)")
}

func SetupCleanupLocalCache(cmdData *SyncCmdData, cmd *cobra.Command) {
	cmdData.CleanupLocalCache = new(bool)
	cmd.Flags().BoolVarP(cmdData.CleanupLocalCache, "cleanup-local-cache", "", common.GetBoolEnvironmentDefaultFalse("WERF_CLEANUP_LOCAL_CACHE"), "Remove intermediate docker images which were created on localhost during sync procedure of two remote stages storages (default $WERF_CLEANUP_LOCAL_CACHE)")
}

func SetupFromStagesStorage(commonCmdData *common.CmdData, cmdData *SyncCmdData, cmd *cobra.Command) {
	if commonCmdData.CommonRepoData == nil {
		common.SetupCommonRepoData(commonCmdData, cmd)
	}

	cmdData.FromStagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.FromStagesStorage, "from", "", os.Getenv("WERF_FROM"), fmt.Sprintf("Source stages storage from which stages will be moved (docker repo address or :local should be specified, default $WERF_FROM environment)."))

	cmdData.FromStagesStorageRepoData = &common.RepoData{DesignationStorageName: "source stages storage"}

	common.SetupImplementationForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-implementation", []string{"WERF_FROM_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	common.SetupDockerHubUsernameForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-docker-hub-username", []string{"WERF_FROM_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	common.SetupDockerHubPasswordForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-docker-hub-password", []string{"WERF_FROM_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	common.SetupDockerHubTokenForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-docker-hub-token", []string{"WERF_FROM_REPO_DOCKER_HUB_TOKEN", "WERF_REPO_DOCKER_HUB_TOKEN"})
	common.SetupGithubTokenForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-github-token", []string{"WERF_FROM_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	common.SetupHarborUsernameForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-harbor-username", []string{"WERF_FROM_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	common.SetupHarborPasswordForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-harbor-password", []string{"WERF_FROM_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
	common.SetupQuayTokenForRepoData(cmdData.FromStagesStorageRepoData, cmd, "from-repo-quay-token", []string{"WERF_FROM_REPO_QUAY_TOKEN", "WERF_REPO_QUAY_TOKEN"})
}

func SetupToStagesStorage(commonCmdData *common.CmdData, cmdData *SyncCmdData, cmd *cobra.Command) {
	if commonCmdData.CommonRepoData == nil {
		common.SetupCommonRepoData(commonCmdData, cmd)
	}

	cmdData.ToStagesStorage = new(string)
	cmd.Flags().StringVarP(cmdData.ToStagesStorage, "to", "", os.Getenv("WERF_TO"), fmt.Sprintf("Destination stages storage to which stages will be moved (docker repo address or :local should be specified, default $WERF_TO environment)."))

	cmdData.ToStagesStorageRepoData = &common.RepoData{DesignationStorageName: "destination stages storage"}

	common.SetupImplementationForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-implementation", []string{"WERF_TO_REPO_IMPLEMENTATION", "WERF_REPO_IMPLEMENTATION"})
	common.SetupDockerHubUsernameForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-docker-hub-username", []string{"WERF_TO_REPO_DOCKER_HUB_USERNAME", "WERF_REPO_DOCKER_HUB_USERNAME"})
	common.SetupDockerHubPasswordForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-docker-hub-password", []string{"WERF_TO_REPO_DOCKER_HUB_PASSWORD", "WERF_REPO_DOCKER_HUB_PASSWORD"})
	common.SetupDockerHubTokenForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-docker-hub-token", []string{"WERF_TO_REPO_DOCKER_HUB_TOKEN", "WERF_REPO_DOCKER_HUB_TOKEN"})
	common.SetupGithubTokenForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-github-token", []string{"WERF_TO_REPO_GITHUB_TOKEN", "WERF_REPO_GITHUB_TOKEN"})
	common.SetupHarborUsernameForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-harbor-username", []string{"WERF_TO_REPO_HARBOR_USERNAME", "WERF_REPO_HARBOR_USERNAME"})
	common.SetupHarborPasswordForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-harbor-password", []string{"WERF_TO_REPO_HARBOR_PASSWORD", "WERF_REPO_HARBOR_PASSWORD"})
	common.SetupQuayTokenForRepoData(cmdData.ToStagesStorageRepoData, cmd, "to-repo-quay-token", []string{"WERF_TO_REPO_QUAY_TOKEN", "WERF_REPO_QUAY_TOKEN"})
}

func NewFromStagesStorage(commonCmdData *common.CmdData, cmdData *SyncCmdData, containerRuntime container_runtime.ContainerRuntime, defaultAddress string) (storage.StagesStorage, error) {
	var address string
	if *cmdData.FromStagesStorage == "" {
		if defaultAddress == "" {
			return nil, fmt.Errorf("--from=ADDRESS param required")
		}
		address = defaultAddress
	} else {
		address = *cmdData.FromStagesStorage
	}

	repoData := common.MergeRepoData(cmdData.FromStagesStorageRepoData, commonCmdData.CommonRepoData)

	if err := common.ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return NewStagesStorage(commonCmdData, address, repoData, containerRuntime)
}

func NewToStagesStorage(commonCmdData *common.CmdData, cmdData *SyncCmdData, containerRuntime container_runtime.ContainerRuntime) (storage.StagesStorage, error) {
	if *cmdData.ToStagesStorage == "" {
		return nil, fmt.Errorf("--to=ADDRESS param required")
	}

	repoData := common.MergeRepoData(cmdData.ToStagesStorageRepoData, commonCmdData.CommonRepoData)

	if err := common.ValidateRepoImplementation(*repoData.Implementation); err != nil {
		return nil, err
	}

	return NewStagesStorage(commonCmdData, *cmdData.ToStagesStorage, repoData, containerRuntime)
}

func NewStagesStorage(commonCmdData *common.CmdData, stagesStorageAddress string, repoData *common.RepoData, containerRuntime container_runtime.ContainerRuntime) (storage.StagesStorage, error) {
	return storage.NewStagesStorage(
		stagesStorageAddress,
		containerRuntime,
		storage.StagesStorageOptions{
			RepoStagesStorageOptions: storage.RepoStagesStorageOptions{
				Implementation: *repoData.Implementation,
				DockerRegistryOptions: docker_registry.DockerRegistryOptions{
					InsecureRegistry:      *commonCmdData.InsecureRegistry,
					SkipTlsVerifyRegistry: *commonCmdData.SkipTlsVerifyRegistry,
					DockerHubUsername:     *repoData.DockerHubUsername,
					DockerHubPassword:     *repoData.DockerHubPassword,
					DockerHubToken:        *repoData.DockerHubToken,
					GitHubToken:           *repoData.GitHubToken,
					HarborUsername:        *repoData.HarborUsername,
					HarborPassword:        *repoData.HarborPassword,
					QuayToken:             *repoData.QuayToken,
				},
			},
		},
	)
}
