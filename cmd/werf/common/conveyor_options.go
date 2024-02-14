package common

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/container_backend/thirdparty/platformutil"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/storage"
)

func GetConveyorOptions(ctx context.Context, commonCmdData *CmdData, imagesToProcess build.ImagesToProcess) (build.ConveyorOptions, error) {
	conveyorOptions := build.ConveyorOptions{
		LocalGitRepoVirtualMergeOptions: stage.VirtualMergeOptions{
			VirtualMerge: *commonCmdData.VirtualMerge,
		},
		ImagesToProcess: imagesToProcess,
	}

	if len(commonCmdData.GetPlatform()) > 0 {
		platforms, err := platformutil.NormalizeUserParams(commonCmdData.GetPlatform())
		if err != nil {
			return build.ConveyorOptions{}, fmt.Errorf("unable to normalize platform params %v: %w", commonCmdData.GetPlatform(), err)
		}
		conveyorOptions.TargetPlatforms = platforms
	}

	conveyorOptions.DeferBuildLog = GetDeferredBuildLog(ctx, commonCmdData)

	return conveyorOptions, nil
}

// GetDeferredBuildLog returns true if build log should be catched and printed on error.
// Default rules are follows:
// - If --require-built-images is specified catch log and print on error.
// - Hide log messages if --log-quiet is specified.
// - Print "live" logs by default or if --log-verbose is specified.
func GetDeferredBuildLog(ctx context.Context, commonCmdData *CmdData) bool {
	requireBuiltImage := GetRequireBuiltImages(ctx, commonCmdData)
	isVerbose := logboek.Context(ctx).IsAcceptedLevel(level.Default)
	return requireBuiltImage || !isVerbose
}

func GetConveyorOptionsWithParallel(ctx context.Context, commonCmdData *CmdData, imagesToProcess build.ImagesToProcess, buildStagesOptions build.BuildOptions) (build.ConveyorOptions, error) {
	conveyorOptions, err := GetConveyorOptions(ctx, commonCmdData, imagesToProcess)
	if err != nil {
		return conveyorOptions, err
	}

	conveyorOptions.Parallel = !(buildStagesOptions.ImageBuildOptions.IntrospectAfterError || buildStagesOptions.ImageBuildOptions.IntrospectBeforeError || len(buildStagesOptions.Targets) != 0) && *commonCmdData.Parallel

	parallelTasksLimit, err := GetParallelTasksLimit(commonCmdData)
	if err != nil {
		return conveyorOptions, fmt.Errorf("getting parallel tasks limit failed: %w", err)
	}
	conveyorOptions.ParallelTasksLimit = parallelTasksLimit

	return conveyorOptions, nil
}

func GetShouldBeBuiltOptions(commonCmdData *CmdData, imageNameList []string) (options build.ShouldBeBuiltOptions, err error) {
	customTagFuncList, err := getCustomTagFuncList(getCustomTagOptionValues(commonCmdData), commonCmdData, imageNameList)
	if err != nil {
		return options, err
	}

	options = build.ShouldBeBuiltOptions{CustomTagFuncList: customTagFuncList}
	return options, nil
}

func GetBuildOptions(ctx context.Context, commonCmdData *CmdData, werfConfig *config.WerfConfig, imageNameList []string) (buildOptions build.BuildOptions, err error) {
	introspectOptions, err := GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return buildOptions, err
	}

	customTagFuncList, err := getCustomTagFuncList(getCustomTagOptionValues(commonCmdData), commonCmdData, imageNameList)
	if err != nil {
		return buildOptions, err
	}

	buildOptions = build.BuildOptions{
		SkipImageMetadataPublication: *commonCmdData.Dev || werfConfig.Meta.Cleanup.DisableGitHistoryBasedPolicy,
		CustomTagFuncList:            customTagFuncList,
		ImageBuildOptions: container_backend.BuildOptions{
			IntrospectAfterError:  *commonCmdData.IntrospectAfterError,
			IntrospectBeforeError: *commonCmdData.IntrospectBeforeError,
		},
		IntrospectOptions: introspectOptions,
	}

	usedNewBuildReportOption := (commonCmdData.SaveBuildReport != nil && *commonCmdData.SaveBuildReport == true) || (commonCmdData.BuildReportPath != nil && *commonCmdData.BuildReportPath != "")

	usedOldBuildReportOption := (commonCmdData.DeprecatedReportPath != nil && *commonCmdData.DeprecatedReportPath != "") || (commonCmdData.DeprecatedReportFormat != nil && *commonCmdData.DeprecatedReportFormat != "")

	if usedNewBuildReportOption && usedOldBuildReportOption {
		return buildOptions, fmt.Errorf("you can't use deprecated options --report-path and --report-format along with new options --save-build-report and --build-report-path, use only the latter instead")
	}

	if usedNewBuildReportOption && GetSaveBuildReport(commonCmdData) {
		buildOptions.ReportPath, buildOptions.ReportFormat, err = GetBuildReportPathAndFormat(commonCmdData)
		if err != nil {
			return buildOptions, fmt.Errorf("getting build report path failed: %w", err)
		}
	} else if usedOldBuildReportOption {
		buildOptions.ReportFormat, err = GetDeprecatedReportFormat(ctx, commonCmdData)
		if err != nil {
			return buildOptions, fmt.Errorf("getting deprecated build report format failed: %w", err)
		}

		buildOptions.ReportPath = GetDeprecatedReportPath(ctx, commonCmdData)
	}

	return buildOptions, nil
}

func getCustomTagFuncList(tagOptionValues []string, commonCmdData *CmdData, imageNameList []string) ([]image.CustomTagFunc, error) {
	if len(tagOptionValues) == 0 {
		return nil, nil
	}

	if *commonCmdData.Repo.Address == "" || *commonCmdData.Repo.Address == storage.LocalStorageAddress {
		return nil, fmt.Errorf("custom tags can only be used with remote storage: --repo=ADDRESS param required")
	}

	templateName := "--add/use-custom-tag"
	tmpl := template.New(templateName).Delims("%", "%")
	tmpl = tmpl.Funcs(map[string]interface{}{
		"image":                   func() string { return "%[1]s" },
		"image_slug":              func() string { return "%[2]s" },
		"image_safe_slug":         func() string { return "%[3]s" },
		"image_content_based_tag": func() string { return "%[4]s" },
	})

	var tagFuncList []image.CustomTagFunc
	for _, optionValue := range tagOptionValues {
		tmpl, err := tmpl.Parse(optionValue)
		if err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
		}

		buf := bytes.NewBuffer(nil)
		if err := tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
		}

		tagOrFormat := buf.String()
		tagFunc := func(imageName, contentBasedTag string) string {
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

			if err := slug.ValidateDockerTag(imageTag); err != nil {
				return nil, fmt.Errorf("invalid custom tag %q: %w", optionValue, err)
			}

			if prevImageTag == "" {
				prevImageTag = imageTag
				continue
			} else if imageTag == prevImageTag {
				return nil, fmt.Errorf("invalid custom tag %q: it is necessary to use the image name in the tag format if there is more than one image in the werf config (e.g., %q)", tagOrFormat, fmt.Sprintf("%s-%s", "%image%", tagOrFormat))
			}
		}

		tagFuncList = append(tagFuncList, tagFunc)
	}

	return tagFuncList, nil
}

func GetUseCustomTagFunc(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, imageNameList []string) (image.CustomTagFunc, error) {
	var tagOptionValues []string
	if *commonCmdData.UseCustomTag != "" {
		tagOptionValues = []string{*commonCmdData.UseCustomTag}
	}

	customTagFuncList, err := getCustomTagFuncList(tagOptionValues, commonCmdData, imageNameList)
	if err != nil {
		return nil, err
	}

	if len(customTagFuncList) == 0 {
		return nil, nil
	}

	if err := giterminismManager.Inspector().InspectCustomTags(); err != nil {
		return nil, err
	}

	if len(customTagFuncList) != 1 {
		panic("unexpected condition")
	}

	return customTagFuncList[0], nil
}

func getCustomTagOptionValues(commonCmdData *CmdData) []string {
	var tagOptionValues []string

	if commonCmdData.UseCustomTag != nil && *commonCmdData.UseCustomTag != "" {
		tagOptionValues = append(tagOptionValues, *commonCmdData.UseCustomTag)
	}

	if commonCmdData.AddCustomTag != nil {
		tagOptionValues = append(tagOptionValues, getAddCustomTag(commonCmdData)...)
	}

	return tagOptionValues
}
