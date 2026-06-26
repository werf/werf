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

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/featgate"
	chartcommonutil "github.com/werf/nelm/pkg/helm/pkg/chart/common/util"
	"github.com/werf/nelm/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/pkg/helm/pkg/chart/v2"
	"github.com/werf/nelm/pkg/helm/pkg/cli/values"
	helm "github.com/werf/nelm/pkg/helm/pkg/cmd"
	"github.com/werf/nelm/pkg/helm/pkg/downloader"
	"github.com/werf/nelm/pkg/helm/pkg/getter"
	legacysecret "github.com/werf/nelm/pkg/legacy/secret"
	"github.com/werf/nelm/pkg/ts"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/build"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/deploy"
	"github.com/werf/werf/v2/pkg/deploy/bundles"
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
	common.SetupConfigRenderPath(&commonCmdData, cmd)
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
	common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupContainerRegistryMirror(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	common.SetupDenoBinaryPath(&commonCmdData, cmd)

	common.SetupSaveBuildReport(&commonCmdData, cmd)
	common.SetupBuildReportPath(&commonCmdData, cmd)
	common.SetupUseBuildReport(&commonCmdData, cmd)

	common.SetupUseCustomTag(&commonCmdData, cmd)
	common.SetupAddCustomTag(&commonCmdData, cmd)

	common.SetupParallelOptions(&commonCmdData, cmd, common.DefaultBuildParallelTasksLimit)

	common.SetupDisableAutoHostCleanup(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedBackendStorageVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsage(&commonCmdData, cmd)
	common.SetupAllowedLocalCacheVolumeUsageMargin(&commonCmdData, cmd)
	common.SetupBackendStoragePath(&commonCmdData, cmd)
	common.SetupProjectName(&commonCmdData, cmd, false)

	common.SetupRequireBuiltImages(&commonCmdData, cmd)
	commonCmdData.SetupPlatform(cmd)
	commonCmdData.SetupBackendNetwork(cmd)

	commonCmdData.SetupSkipImageSpecStage(cmd)
	commonCmdData.SetupDebugTemplates(cmd)
	commonCmdData.SetupAllowIncludesUpdate(cmd)

	lo.Must0(common.SetupChartRepoConnectionFlags(&commonCmdData, cmd))
	lo.Must0(common.SetupValuesFlags(&commonCmdData, cmd))
	common.SetupSecretValuesFileFlags(&commonCmdData, cmd)

	common.SetupAddAnnotations(&commonCmdData, cmd)
	common.SetupAddLabels(&commonCmdData, cmd)
	commonCmdData.SetupSkipDependenciesRepoRefresh(cmd)
	commonCmdData.SetupHelmCompatibleChart(cmd, false)
	commonCmdData.SetupRenameChart(cmd)

	defaultTag := os.Getenv("WERF_TAG")
	if defaultTag == "" {
		defaultTag = "latest"
	}
	cmd.Flags().StringVarP(&cmdData.Tag, "tag", "", defaultTag, "Publish bundle into container registry repo by the provided tag ($WERF_TAG or latest by default)")

	return cmd
}

func runPublish(ctx context.Context, imageNameListFromArgs []string) error {
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
	})
	if err != nil {
		return fmt.Errorf("component init error: %w", err)
	}

	containerBackend := commonManager.ContainerBackend()

	defer func() {
		if err := tmp_manager.DelegateCleanup(ctx); err != nil {
			logboek.Context(ctx).Warn().LogF("Temporary files cleanup preparation failed: %s\n", err)
		}
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

	werfConfigPath, werfConfig, err := common.GetRequiredWerfConfig(ctx, &commonCmdData, giterminismManager, config.WerfConfigOptions{LogRenderedFilePath: true, Env: commonCmdData.Environment})
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

		conveyorWithRetry := build.NewConveyorWithRetryWrapper(werfConfig, giterminismManager, giterminismManager.ProjectDir(), projectTmpDir, containerBackend, storageManager, conveyorOptions)
		defer conveyorWithRetry.Terminate()

		if err := conveyorWithRetry.WithRetryBlock(ctx, func(c *build.Conveyor) error {
			if c.UseBuildReport {
				logboek.Context(ctx).Debug().LogFDetails("Avoid building because of using build report: %s\n", c.BuildReportPath)

				imagesInfoGetters, err = c.GetImageInfoGettersFromReport(ctx, image.InfoGetterOptions{CustomTagFunc: useCustomTagFunc})
				if err != nil {
					return err
				}
			} else {
				if common.GetRequireBuiltImages(&commonCmdData) {
					shouldBeBuiltOptions, err := common.GetShouldBeBuiltOptions(&commonCmdData, werfConfig, imagesToProcess)
					if err != nil {
						return err
					}

					if _, err := c.ShouldBeBuilt(ctx, shouldBeBuiltOptions); err != nil {
						return err
					}
				} else {
					if _, err := c.Build(ctx, buildOptions); err != nil {
						return err
					}
				}

				imagesInfoGetters, err = c.GetImageInfoGetters(image.InfoGetterOptions{CustomTagFunc: useCustomTagFunc})
				if err != nil {
					return err
				}
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
	serviceAnnotations["project.werf.io/env"] = commonCmdData.Environment

	extraLabels, err := common.GetUserExtraLabels(&commonCmdData)
	if err != nil {
		return fmt.Errorf("get user extra labels: %w", err)
	}

	helmRegistryClient, err := common.NewHelmRegistryClient(ctx, *commonCmdData.DockerConfig, commonCmdData.ChartRepoInsecure)
	if err != nil {
		return err
	}

	bundlesRegistryClient, err := common.NewBundlesRegistryClient(ctx, &commonCmdData)
	if err != nil {
		return err
	}

	nelmcommon.ChartFileReader = giterminismManager.FileManager

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit hash failed: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return fmt.Errorf("getting HEAD commit time failed: %w", err)
	}

	serviceValues, err := deploy.GetServiceValues(ctx, werfConfig.Meta.Project, imagesRepo, imagesInfoGetters, deploy.ServiceValuesOptions{
		Env:        commonCmdData.Environment,
		CommitHash: headHash,
		CommitDate: headTime,
	})
	if err != nil {
		return fmt.Errorf("get service values: %w", err)
	}

	helm.Settings.Debug = *commonCmdData.LogDebug

	sv, err := bundles.BundleTagToChartVersion(ctx, cmdData.Tag, time.Now())
	if err != nil {
		return fmt.Errorf("unable to set chart version from bundle tag %q: %w", cmdData.Tag, err)
	}
	chartVersion := sv.String()

	bundleTmpDir := filepath.Join(werf.GetServiceDir(), "tmp", "bundles", uuid.NewString())
	defer os.RemoveAll(bundleTmpDir)

	opts := nelmcommon.HelmOptions{
		ChartLoadOpts: nelmcommon.ChartLoadOptions{
			ChartAppVersion:            common.GetHelmChartConfigAppVersion(werfConfig),
			DefaultChartAPIVersion:     chart.APIVersionV2,
			DefaultChartName:           werfConfig.Meta.Project,
			DefaultChartVersion:        "1.0.0",
			DefaultSecretValuesDisable: commonCmdData.DefaultSecretValuesDisable,
			DefaultValuesDisable:       commonCmdData.DefaultValuesDisable,
			NoSecrets:                  true,
			ChartDepsDownloader: &downloader.Manager{
				Out:               logboek.Context(ctx).OutStream(),
				ChartPath:         bundleTmpDir,
				ContentCache:      helm.Settings.ContentCache,
				AllowMissingRepos: true,
				Getters:           getter.Getters(),
				RegistryClient:    helmRegistryClient,
				RepositoryConfig:  helm.Settings.RepositoryConfig,
				RepositoryCache:   helm.Settings.RepositoryCache,
				Debug:             helm.Settings.Debug,
			},
			ExtraValues:       serviceValues,
			SecretValuesFiles: commonCmdData.SecretValuesFiles,
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
		commonCmdData.DenoBinaryPath,
		&values.Options{
			ValueFiles:    commonCmdData.ValuesFiles,
			StringValues:  commonCmdData.ValuesSetString,
			Values:        commonCmdData.ValuesSet,
			FileValues:    commonCmdData.ValuesSetFile,
			JSONValues:    commonCmdData.ValuesSetJSON,
			LiteralValues: commonCmdData.ValuesSetLiteral,
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

	opts.ChartLoadOpts.ChartType = nelmcommon.LegacyChartTypeBundle

	return bundles.Publish(ctx, bundleTmpDir, fmt.Sprintf("%s:%s", bundleRepo, cmdData.Tag), bundlesRegistryClient, bundles.PublishOptions{
		HelmCompatibleChart: commonCmdData.HelmCompatibleChart,
		RenameChart:         commonCmdData.RenameChart,
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
	denoBinaryPath string,
	vals *values.Options,
	opts nelmcommon.HelmOptions,
) error {
	loadedChart, err := loader.LoadDir(nelmcommon.ContextWithHelmOptions(ctx, opts), chartDir)
	if err != nil {
		return fmt.Errorf("error loading chart %q: %w", chartDir, err)
	}

	chrt, ok := loadedChart.(*chart.Chart)
	if !ok {
		return fmt.Errorf("unsupported chart type %T", loadedChart)
	}

	if featgate.FeatGateTypescript.Enabled() {
		if err := ts.BundleChartsRecursive(ctx, chrt, chartDir, true, denoBinaryPath); err != nil {
			return fmt.Errorf("unable to process TypeScript files in chart: %w", err)
		}
	}

	var valsData []byte
	{
		p := getter.Getters()
		vals, err := vals.MergeValues(nelmcommon.ContextWithHelmOptions(ctx, opts), p)
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

	mergedSecretVals, err := mergeRawSecretValues(ctx, chrt, commonCmdData.SecretValuesFiles, commonCmdData.DefaultSecretValuesDisable)
	if err != nil {
		return fmt.Errorf("unable to merge secret values: %w", err)
	}

	var secretValsData []byte
	if len(mergedSecretVals) > 0 {
		secretValsData, err = yaml.Marshal(mergedSecretVals)
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

	ignoreChartValuesFiles := []string{legacysecret.DefaultSecretValuesFileName}

	// Do not publish into the bundle no custom values nor custom secret values.
	// Final bundle values and secret values will be preconstructed, merged and
	//  embedded into the bundle using only 2 files: values.yaml and secret-values.yaml.
	for _, customValuesPath := range append(commonCmdData.SecretValuesFiles, vals.ValueFiles...) {
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

	for _, f := range chrt.RuntimeFiles {
		if err := writeChartFile(ctx, destDir, f.Name, f.Data); err != nil {
			return fmt.Errorf("error writing chart runtime file: %w", err)
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
	vals, err := chartcommonutil.MergeInternal(ctx, inputVals, serviceValues, nil)
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

	mergedVals, err := chartcommonutil.MergeValues(chrt, valsCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to merge chart values: %w", err)
	}

	valsCopy = mergedVals

	return valsCopy, nil
}

func mergeRawSecretValues(ctx context.Context, chrt *chart.Chart, customSecretValuesFiles []string, defaultSecretValuesDisable bool) (map[string]interface{}, error) {
	var result map[string]interface{}

	if !defaultSecretValuesDisable {
		for _, f := range chrt.Files {
			if f.Name == legacysecret.DefaultSecretValuesFileName {
				vals := map[string]interface{}{}
				if err := yaml.Unmarshal(f.Data, &vals); err != nil {
					return nil, fmt.Errorf("unmarshal %s: %w", f.Name, err)
				}
				result = nelmcommon.LegacyCoalesceTablesFunc(vals, result)
				break
			}
		}
	}

	for _, filePath := range customSecretValuesFiles {
		data, err := nelmcommon.ChartFileReader.ReadChartFile(ctx, filePath)
		if err != nil {
			return nil, fmt.Errorf("read secret values file %q: %w", filePath, err)
		}

		vals := map[string]interface{}{}
		if err := yaml.Unmarshal(data, &vals); err != nil {
			return nil, fmt.Errorf("unmarshal secret values file %q: %w", filePath, err)
		}

		result = nelmcommon.LegacyCoalesceTablesFunc(vals, result)
	}

	return result, nil
}
