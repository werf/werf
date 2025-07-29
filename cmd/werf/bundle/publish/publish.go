package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/copystructure"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	helm_v3 "github.com/werf/3p-helm/cmd/helm"
	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/cli/values"
	"github.com/werf/3p-helm/pkg/downloader"
	"github.com/werf/3p-helm/pkg/getter"
	"github.com/werf/3p-helm/pkg/werf/file"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/3p-helm/pkg/werf/secrets"
	"github.com/werf/3p-helm/pkg/werf/secrets/runtimedata"
	"github.com/werf/common-go/pkg/secrets_manager"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
	"github.com/werf/werf/v2/pkg/deploy/helm/chart_extender/helpers"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/tmp_manager"
	"github.com/werf/werf/v2/pkg/true_git"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var cmdData struct {
	Tag string
}

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "publish [IMAGE_NAME...]",
		Short:                 "Publish bundle",
		Long:                  common.GetLongCommandDescription(GetBundlePublishDocs().Long),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			common.CmdEnvAnno: common.EnvsDescription(),
			common.DocsLongMD: GetBundlePublishDocs().LongMD,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			common.LogVersion()

			return common.LogRunningTime(func() error { return runPublish(ctx, args) })
		},
	})

	commonCmdData.SetupWithoutImages(cmd)
	commonCmdData.SetupFinalImagesOnly(cmd, true)

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupGiterminismOptions(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})
	common.SetupSSHKey(&commonCmdData, cmd)

	common.SetupIntrospectAfterError(&commonCmdData, cmd)
	common.SetupIntrospectBeforeError(&commonCmdData, cmd)
	common.SetupIntrospectStage(&commonCmdData, cmd)

	common.SetupSecondaryStagesStorageOptions(&commonCmdData, cmd)
	common.SetupCacheStagesStorageOptions(&commonCmdData, cmd)
	common.SetupRepoOptions(&commonCmdData, cmd, common.RepoDataOptions{})
	common.SetupFinalRepo(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "Command needs granted permissions to read, pull and push images into the specified repo and to pull base images")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	common.SetupInsecureHelmDependencies(&commonCmdData, cmd, true)
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupSynchronization(&commonCmdData, cmd)

	common.SetupKubeConfig(&commonCmdData, cmd)
	common.SetupKubeConfigBase64(&commonCmdData, cmd)
	common.SetupKubeContext(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)

	common.SetupSet(&commonCmdData, cmd)
	common.SetupSetString(&commonCmdData, cmd)
	common.SetupSetFile(&commonCmdData, cmd)
	common.SetupValues(&commonCmdData, cmd, true)
	common.SetupSecretValues(&commonCmdData, cmd, true)
	common.SetupIgnoreSecretKey(&commonCmdData, cmd)

	commonCmdData.SetupDisableDefaultValues(cmd)
	commonCmdData.SetupDisableDefaultSecretValues(cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)
	common.SetupVirtualMerge(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	common.SetupProjectName(&commonCmdData, cmd, false)

	common.SetupAnnotateLayersWithDmVerityRootHash(&commonCmdData, cmd)
	common.SetupSigningOptions(&commonCmdData, cmd)
	common.SetupELFSigningOptions(&commonCmdData, cmd)

	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)

	commonCmdData.SetupHelmCompatibleChart(cmd, false)
	commonCmdData.SetupRenameChart(cmd)

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Publish bundle into container registry repo by the provided tag ($WERF_TAG or latest by default)")

	return cmd
}

func runPublish(ctx context.Context, imageNameListFromArgs []string) error {
	global_warnings.PostponeMultiwerfNotUpToDateWarning(ctx)
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
		InitSSHAgent:                true,

		SetupOndemandKubeInitializer: true,
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		if err := common.RunAutoHostCleanup(ctx, &commonCmdData, containerBackend); err != nil {
			logboek.Context(ctx).Error().LogF("Auto host cleanup failed: %s\n", err)
		}
	}()

	defer func() {
		commonManager.TerminateSSHAgent()
	}()

	giterminismManager, err := common.GetGiterminismManager(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	common.ProcessLogProjectDir(&commonCmdData, giterminismManager.ProjectDir())

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: *commonCmdData.Environment})
	if err != nil {
		return fmt.Errorf("unable to load werf config: %w", err)
	}

	imagesToProcess, err := config.NewImagesToProcess(werfConfig, imageNameListFromArgs, *commonCmdData.FinalImagesOnly, *commonCmdData.WithoutImages)
	if err != nil {
		return err
	}

	projectName := werfConfig.Meta.Project

	projectTmpDir, err := tmp_manager.CreateProjectDir(ctx)
	if err != nil {
		return fmt.Errorf("getting project tmp dir failed: %w", err)
	}
	defer tmp_manager.ReleaseProjectDir(projectTmpDir)

	buildOptions, err := common.GetBuildOptions(ctx, &commonCmdData, werfConfig, imagesToProcess)
	if err != nil {
		return err
	}

	logboek.LogOptionalLn()

	stagesStorage, err := common.GetStagesStorage(ctx, containerBackend, &commonCmdData, common.GetStagesStorageOpts{
		CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
		GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
	})
	if err != nil {
		return err
	}
	finalStagesStorage, err := common.GetOptionalFinalStagesStorage(ctx, containerBackend, &commonCmdData)
	if err != nil {
		return err
	}

	var imagesInfoGetters []*image.InfoGetter
	var imagesRepo string

	if !imagesToProcess.WithoutImages {

		useCustomTagFunc, err := common.GetUseCustomTagFunc(&commonCmdData, giterminismManager, imagesToProcess)
		if err != nil {
			return err
		}

		storageManager, err := common.NewStorageManager(ctx, &common.NewStorageManagerConfig{
			ProjectName:                    projectName,
			ContainerBackend:               containerBackend,
			CmdData:                        &commonCmdData,
			CleanupDisabled:                werfConfig.Meta.Cleanup.DisableCleanup,
			GitHistoryBasedCleanupDisabled: werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		})
		if err != nil {
			return fmt.Errorf("unable to init storage manager: %w", err)
		}

		imagesRepo = storageManager.GetServiceValuesRepo()

		conveyorOptions, err := common.GetConveyorOptionsWithParallel(ctx, &commonCmdData, imagesToProcess, buildOptions)
		if err != nil {
			return err
		}

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, storageManager.StorageLockManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if common.GetRequireBuiltImages(ctx, &commonCmdData) {
				shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, imagesToProcess)
				if err != nil {
					return err
				}

				if err := c.ShouldBeBuilt(ctx, shouldBeBuiltOptions); err != nil {
					return err
				}
			} else {
				if err := c.Build(ctx, buildOptions); err != nil {
					return err
				}
			}

			imagesInfoGetters, err = c.GetImageInfoGetters(image.InfoGetterOptions{CustomTagFunc: useCustomTagFunc})
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		logboek.LogOptionalLn()
	}

	chartDir, err := common.GetHelmChartDir(werfConfigPath, werfConfig, giterminismManager)
	if err != nil {
		return fmt.Errorf("getting helm chart dir failed: %w", err)
	}

	serviceAnnotations := map[string]string{}
	extraAnnotations := map[string]string{}
	if annos, err := common.GetUserExtraAnnotations(&commonCmdData); err != nil {
		return fmt.Errorf("get user extra annotations: %w", err)
	} else {
		for key, value := range annos {
			if strings.HasPrefix(key, "project.werf.io/") ||
				strings.Contains(key, "ci.werf.io/") ||
				key == "werf.io/release-channel" {
				serviceAnnotations[key] = value
			} else {
				extraAnnotations[key] = value
			}
		}
	}

	serviceAnnotations["project.werf.io/name"] = werfConfig.Meta.Project
	serviceAnnotations["project.werf.io/env"] = *commonCmdData.Environment

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, *commonCmdData.InsecureHelmDependencies)
	if err != nil {
		return err
	}

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	file.ChartFileReader = giterminismManager.FileManager

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit hash failed: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit time failed: %w", err)
	}

	serviceValues, err := helpers.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepo, imagesInfoGetters, helpers.ServiceValuesOptions{
		Env:        *commonCmdData.Environment,
		CommitHash: headHash,
		CommitDate: headTime,
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	helm_v3.Settings.Debug = *commonCmdData.LogDebug

	// FIXME(1.3): compatibility mode with older 1.2 versions, which do not require WERF_SECRET_KEY in the 'werf bundle publish' command
	if err := secrets_manager.Manager.AllowMissedSecretKeyMode(giterminismManager.ProjectDir()); err != nil {
		return err
	}

	sv, err := bundles.BundleTagToChartVersion(ctx, cmdData.Tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", cmdData.Tag, err)
	}
	chartVersion := sv.String()

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundleTmpDir)

	opts := helmopts.HelmOptions{
		ChartLoadOpts: helmopts.ChartLoadOptions{
			ChartAppVersion:        common.GetHelmChartConfigAppVersion(werfConfig),
			DefaultChartAPIVersion: chart.APIVersionV2,
			DefaultChartName:       werfConfig.Meta.Project,
			DefaultChartVersion:    "1.0.0",
			DepDownloader: &downloader.Manager{
				Out:               logboek.Context(ctx).OutStream(),
				ChartPath:         bundleTmpDir,
				AllowMissingRepos: true,
				Getters:           getter.All(helm_v3.Settings),
				RegistryClient:    helmRegistryClient,
				RepositoryConfig:  helm_v3.Settings.RepositoryConfig,
				RepositoryCache:   helm_v3.Settings.RepositoryCache,
				Debug:             helm_v3.Settings.Debug,
			},
			NoDecryptSecrets:      *commonCmdData.IgnoreSecretKey,
			NoDefaultSecretValues: *commonCmdData.DisableDefaultSecretValues,
			NoDefaultValues:       *commonCmdData.DisableDefaultValues,
			SecretValuesFiles:     common.GetSecretValues(&commonCmdData),
			SecretsWorkingDir:     giterminismManager.ProjectDir(),
			ExtraValues:           serviceValues,
		},
	}

	if err = createNewBundle(
		ctx,
		serviceValues,
		extraAnnotations,
		serviceAnnotations,
		extraLabels,
		giterminismManager.ProjectDir(),
		chartDir,
		bundleTmpDir,
		chartVersion,
		&values.Options{
			ValueFiles:   common.GetValues(&commonCmdData),
			StringValues: common.GetSetString(&commonCmdData),
			Values:       common.GetSet(&commonCmdData),
			FileValues:   common.GetSetFile(&commonCmdData),
		},
		opts,
	); err != nil {
		return fmt.Errorf("create bundle: %w", err)
	}

	var bundleRepo string
	if finalStagesStorage != nil {
		bundleRepo = finalStagesStorage.Address()
	} else {
		bundleRepo = stagesStorage.Address()
	}

	opts.ChartLoadOpts.ChartType = helmopts.ChartTypeBundle

	return bundles.Publish(ctx, bundleTmpDir, fmt.Sprintf("%s:%s", bundleRepo, cmdData.Tag), bundlesRegistryClient, bundles.PublishOptions{
		HelmCompatibleChart: *commonCmdData.HelmCompatibleChart,
		RenameChart:         *commonCmdData.RenameChart,
		HelmOptions:         opts,
	})
}

func createNewBundle(
	ctx context.Context,
	serviceValues map[string]interface{},
	extraAnnotations map[string]string,
	serviceAnnotations map[string]string,
	extraLabels map[string]string,
	projectDir string,
	chartDir string,
	destDir string,
	chartVersion string,
	vals *values.Options,
	opts helmopts.HelmOptions,
) error {
	chrt, err := loader.LoadDir(chartDir, opts)
	if err != nil {
		return fmt.Errorf("error loading chart %q: %w", chartDir, err)
	}

	var valsData []byte
	{
		p := getter.All(helm_v3.Settings)
		vals, err := vals.MergeValues(p, opts)
		if err != nil {
			return fmt.Errorf("unable to merge input values: %w", err)
		}

		bundleVals, err := makeBundleValues(ctx, chrt, vals, serviceValues)
		if err != nil {
			return fmt.Errorf("unable to construct bundle values: %w", err)
		}

		valsData, err = yaml.Marshal(bundleVals)
		if err != nil {
			return fmt.Errorf("unable to marshal bundle values: %w", err)
		}
	}

	var secretValsData []byte
	if chrt.SecretsRuntimeData != nil && !secrets_manager.Manager.IsMissedSecretKeyModeEnabled() {
		vals, err := makeBundleSecretValues(ctx, chrt.SecretsRuntimeData, opts)
		if err != nil {
			return fmt.Errorf("unable to construct bundle secret values: %w", err)
		}

		secretValsData, err = yaml.Marshal(vals)
		if err != nil {
			return fmt.Errorf("unable to marshal bundle secret values: %w", err)
		}
	}

	if destDir == "" {
		destDir = chrt.Metadata.Name
	}

	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("unable to remove %q: %w", destDir, err)
	}
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", destDir, err)
	}

	logboek.Context(ctx).Debug().LogF("Saving bundle values:\n%s\n---\n", valsData)

	valuesFile := filepath.Join(destDir, "values.yaml")
	if err := ioutil.WriteFile(valuesFile, valsData, os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", valuesFile, err)
	}

	if secretValsData != nil {
		secretValuesFile := filepath.Join(destDir, "secret-values.yaml")
		if err := ioutil.WriteFile(secretValuesFile, secretValsData, os.ModePerm); err != nil {
			return fmt.Errorf("unable to write %q: %w", secretValuesFile, err)
		}
	}

	if chrt.Metadata == nil {
		panic("unexpected condition")
	}

	bundleMetadata := *chrt.Metadata
	// Force api v2
	bundleMetadata.APIVersion = chart.APIVersionV2
	bundleMetadata.Version = chartVersion

	chartYamlFile := filepath.Join(destDir, "Chart.yaml")
	if data, err := json.Marshal(bundleMetadata); err != nil {
		return fmt.Errorf("unable to prepare Chart.yaml data: %w", err)
	} else if err := ioutil.WriteFile(chartYamlFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", chartYamlFile, err)
	}

	if chrt.Lock != nil {
		chartLockFile := filepath.Join(destDir, "Chart.lock")
		if data, err := json.Marshal(chrt.Lock); err != nil {
			return fmt.Errorf("unable to prepare Chart.lock data: %w", err)
		} else if err := ioutil.WriteFile(chartLockFile, append(data, []byte("\n")...), os.ModePerm); err != nil {
			return fmt.Errorf("unable to write %q: %w", chartLockFile, err)
		}
	}

	templatesDir := filepath.Join(destDir, "templates")
	if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", templatesDir, err)
	}

	for _, f := range chrt.Templates {
		if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
			return fmt.Errorf("error writing chart template: %w", err)
		}
	}

	chartDirAbs := filepath.Join(projectDir, chartDir)

	ignoreChartValuesFiles := []string{secrets.DefaultSecretValuesFileName}

	// Do not publish into the bundle no custom values nor custom secret values.
	// Final bundle values and secret values will be preconstructed, merged and
	//  embedded into the bundle using only 2 files: values.yaml and secret-values.yaml.
	for _, customValuesPath := range append(common.GetSecretValues(&commonCmdData), vals.ValueFiles...) {
		path := util.GetAbsoluteFilepath(customValuesPath)
		if util.IsSubpathOfBasePath(chartDirAbs, path) {
			ignoreChartValuesFiles = append(ignoreChartValuesFiles, util.GetRelativeToBaseFilepath(chartDirAbs, path))
		}
	}

WritingFiles:
	for _, f := range chrt.Files {
		for _, ignoreValuesFile := range ignoreChartValuesFiles {
			if f.Name == ignoreValuesFile {
				continue WritingFiles
			}
		}

		if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
			return fmt.Errorf("error writing miscellaneous chart file: %w", err)
		}
	}

	for _, dep := range chrt.Metadata.Dependencies {
		var depPath string

		switch {
		case dep.Repository == "":
			depPath = filepath.Join("charts", dep.Name)
		case strings.HasPrefix(dep.Repository, "file://"):
			depPath = strings.TrimPrefix(dep.Repository, "file://")
		default:
			depPath = fmt.Sprintf("charts/%s-%s.tgz", dep.Name, dep.Version)
		}

		for _, f := range chrt.Raw {
			if strings.HasPrefix(f.Name, depPath) {
				if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
					return fmt.Errorf("error writing subchart file: %w", err)
				}
			}
		}
	}

	if chrt.Schema != nil {
		schemaFile := filepath.Join(destDir, "values.schema.json")
		if err := writeChartFile(ctx, destDir, "values.schema.json", chrt.Schema); err != nil {
			return fmt.Errorf("error writing chart values schema: %w", err)
		}
		if err := ioutil.WriteFile(schemaFile, chrt.Schema, os.ModePerm); err != nil {
			return fmt.Errorf("unable to write %q: %w", schemaFile, err)
		}
	}

	annotations := lo.Assign(extraAnnotations, serviceAnnotations)
	if len(annotations) > 0 {
		if err := writeBundleJsonMap(annotations, filepath.Join(destDir, "extra_annotations.json")); err != nil {
			return err
		}
	}

	if len(extraLabels) > 0 {
		if err := writeBundleJsonMap(extraLabels, filepath.Join(destDir, "extra_labels.json")); err != nil {
			return err
		}
	}

	return nil
}

func writeChartFile(ctx context.Context, destDir, fileName string, fileData []byte) error {
	p := filepath.Join(destDir, fileName)
	dir := filepath.Dir(p)

	logboek.Context(ctx).Debug().LogF("Writing chart file %q\n", p)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating dir %q: %w", dir, err)
	}
	if err := ioutil.WriteFile(p, fileData, os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", p, err)
	}
	return nil
}

func writeBundleJsonMap(dataMap map[string]string, path string) error {
	if data, err := json.Marshal(dataMap); err != nil {
		return fmt.Errorf("unable to prepare %q data: %w", path, err)
	} else if err := ioutil.WriteFile(path, append(data, []byte("\n")...), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write %q: %w", path, err)
	} else {
		return nil
	}
}

func makeBundleValues(ctx context.Context, chrt *chart.Chart, inputVals, serviceValues map[string]interface{}) (map[string]interface{}, error) {
	vals, err := chartutil.MergeInternal(ctx, inputVals, serviceValues, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to coalesce werf chart values: %w", err)
	}

	v, err := copystructure.Copy(vals)
	if err != nil {
		return vals, err
	}

	valsCopy := v.(map[string]interface{})
	// if we have an empty map, make sure it is initialized
	if valsCopy == nil {
		valsCopy = make(map[string]interface{})
	}

	chartutil.CoalesceChartValues(chrt, valsCopy, true)

	return valsCopy, nil
}

func makeBundleSecretValues(
	ctx context.Context,
	secretsRuntimeData runtimedata.RuntimeData,
	opts helmopts.HelmOptions,
) (map[string]interface{}, error) {
	return secretsRuntimeData.GetEncodedSecretValues(ctx, secrets_manager.Manager, opts.ChartLoadOpts.SecretsWorkingDir, opts.ChartLoadOpts.NoDecryptSecrets)
}
