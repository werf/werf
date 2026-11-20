package info

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/cmd/werf/common"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/true_git"
)

const (
	outputFormatText = "text"
	outputFormatJSON = "json"
)

var cmdData struct {
	OutputFormat string
}

var commonCmdData common.CmdData

type metaRepoInfo struct {
	Project          string `json:"project"`
	Repo             string `json:"repo"`
	SafeguardEnabled bool   `json:"safeguardEnabled"`
	MetaRepo         string `json:"metaRepo"`
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "info",
		DisableFlagsInUseLine: true,
		Short:                 "Show whether --repo is bound to a meta-repo for the project and to which one.",
		Long:                  "Read the per-project meta-repo marker from --repo and print the meta-repo address it records and whether the safeguard is enabled. The command only reads --repo and changes nothing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if cmdData.OutputFormat != outputFormatText && cmdData.OutputFormat != outputFormatJSON {
				return fmt.Errorf("--output-format must be %q or %q, got %q", outputFormatText, outputFormatJSON, cmdData.OutputFormat)
			}

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if cmdData.OutputFormat == outputFormatJSON {
				// stdout carries only the JSON document.
				ctx = logboek.NewContext(ctx, logboek.Context(ctx).NewSubLogger(os.Stderr, os.Stderr))
			}

			return run(ctx)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigRenderPath(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read images from --repo")
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	cmd.Flags().StringVarP(&cmdData.OutputFormat, "output-format", "", cmp.Or(os.Getenv("WERF_OUTPUT_FORMAT"), outputFormatText), "Output format. Options: text, json (default $WERF_OUTPUT_FORMAT or \""+outputFormatText+"\")")

	return cmd
}

func run(ctx context.Context) error {
	commonManager, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
		Cmd: &commonCmdData,
		InitTrueGitWithOptions: &common.InitTrueGitOptions{
			Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
		},
		InitDockerRegistry:          true,
		InitProcessContainerBackend: true,
		InitWerf:                    true,
		InitGitDataManager:          true,
		InitManifestCache:           true,
		InitLRUImagesCache:          true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
	}()

	if _, err = tmp_manager.CreateProjectDir(ctx); err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	_, werfConfig, err := common.GetOptionalWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if werfConfig == nil {
		return fmt.Errorf("run command in the project directory with werf.yaml")
	}
	projectName := werfConfig.Meta.Project

	// The storage manager is not used: it registers a synchronization client in --repo.
	stagesStorage, err := common.GetStagesStorage(ctx, commonManager.ContainerBackend(), &commonCmdData, common.GetStagesStorageOpts{SkipMetaCheck: true})
	if err != nil {
		return fmt.Errorf("get stages storage: %w", err)
	}

	repo, ok := stagesStorage.(*storage.RepoStagesStorage)
	if !ok {
		return fmt.Errorf("--repo must be a container registry address")
	}

	metaRepoAddress, found, err := repo.InspectMetaRepoMarker(ctx, projectName)
	if err != nil {
		return fmt.Errorf("read meta-repo marker: %w", err)
	}
	if found && metaRepoAddress == "" {
		return fmt.Errorf("the meta-repo marker for project %q in %s is malformed (no meta-repo address); run 'werf meta-repo detach' to remove it", projectName, repo.Address())
	}

	return printInfo(metaRepoInfo{
		Project:          projectName,
		Repo:             repo.Address(),
		SafeguardEnabled: found,
		MetaRepo:         metaRepoAddress,
	})
}

func printInfo(info metaRepoInfo) error {
	if cmdData.OutputFormat == outputFormatJSON {
		if err := json.NewEncoder(os.Stdout).Encode(info); err != nil {
			return fmt.Errorf("encode meta-repo info: %w", err)
		}
		return nil
	}

	if !info.SafeguardEnabled {
		fmt.Println("Meta-repo path: not set")
		fmt.Println("Safeguard status: disabled")
		return nil
	}

	fmt.Printf("Meta-repo path: %s\n", info.MetaRepo)
	fmt.Println("Safeguard status: enabled")
	return nil
}
