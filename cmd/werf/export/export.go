package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/git_repo/gitdata"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/ssh_agent"
	"github.com/werf/werf/pkg/storage/lrumeta"
	"github.com/werf/werf/pkg/storage/manager"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/true_git"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewExportCmd(ctx context.Context) *cobra.Command {
	var tagTemplateList []string
	var addLabelArray []string

	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "export [IMAGE_NAME...] [options]",
		Short:                 "Export images",
		Long:                  common.GetLongCommandDescription(GetExportDocs().Long),
		DisableFlagsInUseLine: true,
		Example: `  # Export images to Docker Hub and GitHub container registry
  $ werf export \
      --tag index.docker.io/company/project:%image%-latest \
      --tag ghcr.io/company/project/%image%:latest

  # Export images with extra labels
  $ werf export \
      --tag registry.werf.io/company/project/%image%:latest \
      --add-label io.artifacthub.package.readme-url=https://raw.githubusercontent.com/werf/werf/main/README.md \
      --add-label org.opencontainers.image.created=2023-03-13T11:55:24Z \
      --add-label org.opencontainers.image.description="Official image to run werf in containers"`,
		Annotations: map[string]string{
			common.DisableOptionsInUseLineAnno: "1",
			common.DocsLongMD:                  GetExportDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if len(tagTemplateList) == 0 {
				common.PrintHelp(cmd)
				return fmt.Errorf("required at least one tag template: use the --tag option to specify templates")
			}

			var addLabelMap map[string]string
			var err error
			{
				addLabelArray := append(util.PredefinedValuesByEnvNamePrefix("WERF_EXPORT_ADD_LABEL_"), addLabelArray...)
				addLabelMap, err = common.KeyValueArrayToMap(addLabelArray, "=")
				if err != nil {
					common.PrintHelp(cmd)
					return fmt.Errorf("unsupported --add-label value: %w", err)
				}
			}

			return run(ctx, common.GetImagesToProcess(args, false), tagTemplateList, addLabelMap)
		},
	})

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{OptionalRepo: true})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupSkipBuild(&commonCmdData, cmd)
	common.SetupRequireBuiltImages(&commonCmdData, cmd)

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

	commonCmdData.SetupPlatform(cmd)

	cmd.Flags().StringArrayVarP(&tagTemplateList, "tag", "", []string{}, `Set a tag template (can specify multiple).
It is necessary to use image name shortcut %image% or %image_slug% if multiple images are exported (e.g. REPO:TAG-%image% or REPO-%image%:TAG)`)

	cmd.Flags().StringArrayVarP(&addLabelArray, "add-label", "", []string{}, `Add label to exported images (can specify multiple).
Format: labelName=labelValue.
Also, can be specified with $WERF_EXPORT_ADD_LABEL_* (e.g. $WERF_EXPORT_ADD_LABEL_1=labelName1=labelValue1, $WERF_EXPORT_ADD_LABEL_2=labelName2=labelValue2)`)

	return cmd
}

func run(ctx context.Context, imagesToProcess build.ImagesToProcess, tagTemplateList []string, extraLabels map[string]string) error {
	if imagesToProcess.WithoutImages {
		return nil
	}

	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}

	containerBackend, processCtx, err := common.InitProcessContainerBackend(ctx, &commonCmdData)
	if err != nil {
		return err
	}
	ctx = processCtx

	gitDataManager, err := gitdata.GetHostGitDataManager(ctx)
	if err != nil {
		return fmt.Errorf("error getting host git data manager: %w", err)
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

	if err := true_git.Init(ctx, true_git.Options{LiveGitOutput: *commonCmdData.LogDebug}); err != nil {
		return err
	}

	if err := common.DockerRegistryInit(ctx, &commonCmdData); err != nil {
		return err
	}

	if err := ssh_agent.Init(ctx, common.GetSSHKey(&commonCmdData)); err != nil {
		return fmt.Errorf("cannot initialize ssh agent: %w", err)
	}
	defer func() {
		err := ssh_agent.Terminate()
		if err != nil {
			logboek.Warn().LogF("WARNING: ssh agent termination failed: %s\n", err)
		}
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	_, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, common.GetWerfConfigOptions(&commonCmdData, false))
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}
	if err := werfConfig.CheckThatImagesExist(imagesToProcess.OnlyImages); err != nil {
		return err
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	synchronization, err := common.GetSynchronization(ctx, &commonCmdData, projectName, stagesStorage)
	if err != nil {
		return err
	}
	storageLockManager, err := common.GetStorageLockManager(ctx, synchronization)
	if err != nil {
		return err
	}
	secondaryStagesStorageList, err := common.GetSecondaryStagesStorageList(ctx, stagesStorage, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}
	cacheStagesStorageList, err := common.GetCacheStagesStorageList(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}

	storageManager := manager.NewStorageManager(projectName, stagesStorage, finalStagesStorage, secondaryStagesStorageList, cacheStagesStorageList, storageLockManager)

	logboek.Context(ctx).Info().LogOptionalLn()

	conveyorOptions, err := common.GetConveyorOptions(ctx, &commonCmdData, imagesToProcess)
	if err != nil {
		return err
	}

	conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, ssh_agent.SSHAuthSock, containerBackend, storageManager, storageLockManager, conveyorOptions)
	defer conveyorWithRetry.Terminate()

	return conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
		imageNameList := common.GetImageNameList(imagesToProcess, werfConfig)

		tagFuncList, err := getTagFuncList(imageNameList, tagTemplateList)
		if err != nil {
			return err
		}

		if common.GetRequireBuiltImages(ctx, &commonCmdData) {
			if err := c.ShouldBeBuilt(ctx, build.ShouldBeBuiltOptions{}); err != nil {
				return err
			}
		} else {
			if err := c.Build(ctx, build.BuildOptions{SkipImageMetadataPublication: *commonCmdData.Dev}); err != nil {
				return err
			}
		}

		return c.Export(ctx, build.ExportOptions{
			ExportImageNameList: imageNameList,
			ExportTagFuncList:   tagFuncList,
			MutateConfigFunc: func(config v1.Config) (v1.Config, error) {
				for k, v := range extraLabels {
					config.Labels[k] = v
				}
				return config, nil
			},
		})
	})
}

func getTagFuncList(imageNameList, tagTemplateList []string) ([]image.ExportTagFunc, error) {
	templateName := "--tag"
	tmpl := template.New(templateName).Delims("%", "%")
	tmpl = tmpl.Funcs(map[string]interface{}{
		"image":                   func() string { return "%[1]s" },
		"image_slug":              func() string { return "%[2]s" },
		"image_safe_slug":         func() string { return "%[3]s" },
		"image_content_based_tag": func() string { return "%[4]s" },
	})

	var tagFuncList []image.ExportTagFunc
	for _, tagTemplate := range tagTemplateList {
		tagFunc, err := getExportTagFunc(tmpl, templateName, imageNameList, tagTemplate)
		if err != nil {
			return nil, fmt.Errorf("invalid tag template %q: %w", tagTemplate, err)
		}

		tagFuncList = append(tagFuncList, tagFunc)
	}

	return tagFuncList, nil
}

func getExportTagFunc(tmpl *template.Template, templateName string, imageNameList []string, tagTemplate string) (image.ExportTagFunc, error) {
	tmpl, err := tmpl.Parse(tagTemplate)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err = tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
		return nil, err
	}

	tagOrFormat := buf.String()
	var tagFunc image.ExportTagFunc
	tagFunc = func(imageName, contentBasedTag string) string {
		if strings.ContainsRune(tagOrFormat, '%') {
			return fmt.Sprintf(tagOrFormat, imageName, slug.Slug(imageName), slug.DockerTag(imageName), contentBasedTag)
		} else {
			return tagOrFormat
		}
	}

	contentBasedTagStub := strings.Repeat("x", 70) // 1b77754d35b0a3e603731828ee6f2400c4f937382874db2566c616bb-1624991915332
	var prevImageTag string
	for _, imageName := range imageNameList {
		imageTag := tagFunc(imageName, contentBasedTagStub)

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
