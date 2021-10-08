package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewExportCmd() *cobra.Command {
	var tagTemplateList []string

	cmd := &cobra.Command{
		Use:   "export [IMAGE_NAME...] [options]",
		Short: "Export images",
		Long: common.GetLongCommandDescription(`Export images to an arbitrary repository according to a template specified by the --tag option (build if needed).
All meta-information related to werf is removed from the exported images, and then images are completely under the user's responsibility`),
		DisableFlagsInUseLine: true,
		Example: `  # Export images to Docker Hub and GitHub Container Registry
  $ werf export --tag=index.docker.io/company/project:%image%-latest --tag=ghcr.io/company/project/%image%:latest`,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := common.BackgroundContext()
			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if len(tagTemplateList) == 0 {
				common.PrintHelp(cmd)
				return fmt.Errorf("required at least one tag template: use the --tag option to specify templates")
			}

			return run(ctx, args, tagTemplateList)
		},
	}

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupStagesStorageOptions(&commonCmdData, cmd)
	common.SetupFinalStagesStorageOptions(&commonCmdData, cmd)

	common.SetupSkipBuild(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read and pull images from the specified repo")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)
	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupDryRun(&commonCmdData, cmd)

	common.SetupVirtualMerge(&commonCmdData, cmd)
	common.SetupVirtualMergeFromCommit(&commonCmdData, cmd)
	common.SetupVirtualMergeIntoCommit(&commonCmdData, cmd)

	common.SetupPlatform(&commonCmdData, cmd)

	cmd.Flags().StringArrayVarP(&tagTemplateList, "tag", "", []string{}, `Set a tag template (can specify multiple).
It is necessary to use image name shortcut %image% or %image_slug% if multiple images are exported (e.g. REPO:TAG-%image% or REPO-%image%:TAG)`)

	return cmd
}

func run(ctx context.Context, imagesToProcess, tagTemplateList []string) error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	containerRuntime, processCtx, err := common.InitProcessContainerRuntime(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %s", err)
	}

	if err := git_repo.Init(gitDataManager); err != nil {
		return err
	}

	if err := image.Init(); err != nil {
		return err
	}

	if err := lrumeta.Init(); err != nil {
		return err
	}

	if err := true_git.Init(true_git.Options{LiveGitOutput: *commonCmdData.LogVerbose || *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	giterminismManager, err := common.GetGiterminismManager(&commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %s", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %s", err)
	}

	for _, imageToProcess := range imagesToProcess {
		if !werfConfig.HasImageOrArtifact(imageToProcess) {
			return fmt.Errorf("specified image %s is not found in werf.yaml", logging.ImageLogName(imageToProcess, false))
		}
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %s", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	stagesStorageAddress := common.GetOptionalStagesStorageAddress(&commonCmdData)
	stagesStorage, err := common.GetStagesStorage(stagesStorageAddress, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
	if err != nil {
		return err
	}
	stagesStorageCache, err := common.GetStagesStorageCache(synchronization)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(stagesStorage, containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}
	cacheStagesStorageList, err := common.GetCacheStagesStorageList(containerRuntime, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager, stagesStorageCache)

	logboek.Context(ctx).Info().LogOptionalLn()

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, imagesToProcess, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerRuntime, storageManager, storageLockManager, common.GetConveyorOptions(&commonCmdData))
	defer conveyorWithRetry.Terminate()

	return conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		if len(imagesToProcess) == 0 {
			for _, img := range werfConfig.GetAllImages() {
				imagesToProcess = append(imagesToProcess, img.GetName())
			}
		}

		tagFuncList, err := getTagFuncList(imagesToProcess, tagTemplateList)
		if err != nil {
			return err
		}

		return c.Export(ctx, build.ExportOptions{
			BuildPhaseOptions: build.BuildPhaseOptions{
				BuildOptions:      build.BuildOptions{},
				ShouldBeBuiltMode: *commonCmdData.SkipBuild,
			},
			ExportPhaseOptions: build.ExportPhaseOptions{
				ExportTagFuncList: tagFuncList,
			},
		})
	})
}

func getTagFuncList(imageNameList, tagTemplateList []string) ([]func(string) string, error) {
	templateName := "--tag"
	tmpl := template.New(templateName).Delims("%", "%")
	tmpl = tmpl.Funcs(map[string]interface{}{
		"image":           func() string { return "%[1]s" },
		"image_slug":      func() string { return "%[2]s" },
		"image_safe_slug": func() string { return "%[3]s" },
	})

	var tagFuncList []func(string) string
	for _, tagTemplate := range tagTemplateList {
		tagFunc, err := getExportTagFunc(tmpl, templateName, imageNameList, tagTemplate)
		if err != nil {
			return nil, fmt.Errorf("invalid tag template %q: %s", tagTemplate, err)
		}

		tagFuncList = append(tagFuncList, tagFunc)
	}

	return tagFuncList, nil
}

func getExportTagFunc(tmpl *template.Template, templateName string, imageNameList []string, tagTemplate string) (func(imageName string) string, error) {
	tmpl, err := tmpl.Parse(tagTemplate)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err = tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
		return nil, err
	}

	tagOrFormat := buf.String()
	tagFunc := func(imageName string) string {
		if strings.ContainsRune(tagOrFormat, '%') {
			return fmt.Sprintf(tagOrFormat, imageName, slug.Slug(imageName), slug.DockerTag(imageName))
		} else {
			return tagOrFormat
		}
	}

	var prevImageTag string
	for _, imageName := range imageNameList {
		imageTag := tagFunc(imageName)

		ref, err := name.ParseReference(imageTag, name.WeakValidation)
		if err != nil {
			return nil, err
		}

		if ref.Context().RegistryStr() == name.DefaultRegistry && !strings.HasPrefix(imageTag, name.DefaultRegistry) {
			return nil, errors.New(`
- the command exports images to the registry (cannot export them locally)
- the user must explicitly provide the address "index.docker.io" when using Docker Hub as a registry`)
		}

		if prevImageTag == "" {
			prevImageTag = imageTag
			continue
		} else if imageTag == prevImageTag {
			return nil, errors.New(`tag template must contain image name shortcut %image% or %image_slug% if multiple images are exported (e.g. REPO:TAG-%image% or REPO-%image%:TAG)`)
		}
	}

	return tagFunc, nil
}
