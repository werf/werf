package common

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/storage"
)

func GetConveyorOptions(commonCmdData *CmdData) build.ConveyorOptions {
	return build.ConveyorOptions{
		LocalGitRepoVirtualMergeOptions: stage.VirtualMergeOptions{
			VirtualMerge: *commonCmdData.VirtualMerge,
		},
	}
}

func GetConveyorOptionsWithParallel(commonCmdData *CmdData, buildStagesOptions build.BuildOptions) (build.ConveyorOptions, error) {
	conveyorOptions := GetConveyorOptions(commonCmdData)
	conveyorOptions.Parallel = !(buildStagesOptions.ImageBuildOptions.IntrospectAfterError || buildStagesOptions.ImageBuildOptions.IntrospectBeforeError || len(buildStagesOptions.Targets) != 0) && *commonCmdData.Parallel

	parallelTasksLimit, err := GetParallelTasksLimit(commonCmdData)
	if err != nil {
		return conveyorOptions, fmt.Errorf("getting parallel tasks limit failed: %s", err)
	}

	conveyorOptions.ParallelTasksLimit = parallelTasksLimit

	return conveyorOptions, nil
}

func GetShouldBeBuiltOptions(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, werfConfig *config.WerfConfig) (options build.ShouldBeBuiltOptions, err error) {
	customTagFuncList, err := getCustomTagFuncList(commonCmdData, giterminismManager, werfConfig)
	if err != nil {
		return options, err
	}

	options = build.ShouldBeBuiltOptions{CustomTagFuncList: customTagFuncList}
	return options, nil
}

func GetBuildOptions(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, werfConfig *config.WerfConfig) (buildOptions build.BuildOptions, err error) {
	introspectOptions, err := GetIntrospectOptions(commonCmdData, werfConfig)
	if err != nil {
		return buildOptions, err
	}

	reportFormat, err := GetReportFormat(commonCmdData)
	if err != nil {
		return buildOptions, err
	}

	customTagFuncList, err := getCustomTagFuncList(commonCmdData, giterminismManager, werfConfig)
	if err != nil {
		return buildOptions, err
	}

	buildOptions = build.BuildOptions{
		CustomTagFuncList: customTagFuncList,
		ImageBuildOptions: container_runtime.LegacyBuildOptions{
			IntrospectAfterError:  *commonCmdData.IntrospectAfterError,
			IntrospectBeforeError: *commonCmdData.IntrospectBeforeError,
		},
		IntrospectOptions: introspectOptions,
		ReportPath:        *commonCmdData.ReportPath,
		ReportFormat:      reportFormat,
	}

	return buildOptions, nil
}

func getCustomTagFuncList(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, werfConfig *config.WerfConfig) ([]func(string) string, error) {
	tagOptionValues := getCustomTagOptionValues(commonCmdData)
	if len(tagOptionValues) == 0 {
		return nil, nil
	}

	if *commonCmdData.StagesStorage == "" || *commonCmdData.StagesStorage == storage.LocalStorageAddress {
		return nil, fmt.Errorf("custom tags can only be used with remote storage: --repo=ADDRESS param required")
	}

	if err := giterminismManager.Inspector().InspectCustomTags(); err != nil {
		return nil, err
	}

	templateName := "--add/use-custom-tag"
	tmpl := template.New(templateName).Delims("%", "%")
	tmpl = tmpl.Funcs(map[string]interface{}{
		"image":           func() string { return "%[1]s" },
		"image_slug":      func() string { return "%[2]s" },
		"image_safe_slug": func() string { return "%[3]s" },
	})

	var tagFuncList []func(string) string
	for _, optionValue := range tagOptionValues {
		tmpl, err := tmpl.Parse(optionValue)
		if err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %s", optionValue, err)
		}

		buf := bytes.NewBuffer(nil)
		if err := tmpl.ExecuteTemplate(buf, templateName, nil); err != nil {
			return nil, fmt.Errorf("invalid custom tag %q: %s", optionValue, err)
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
		for _, img := range werfConfig.GetAllImages() {
			imageTag := tagFunc(img.GetName())

			if err := slug.ValidateDockerTag(imageTag); err != nil {
				return nil, fmt.Errorf("invalid custom tag %q: %s", optionValue, err)
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

func GetUseCustomTagFunc(commonCmdData *CmdData, giterminismManager giterminism_manager.Interface, werfConfig *config.WerfConfig) (func(string) string, error) {
	customTagFuncList, err := getCustomTagFuncList(commonCmdData, giterminismManager, werfConfig)
	if err != nil {
		return nil, err
	}

	if len(customTagFuncList) == 0 {
		return nil, nil
	}

	if len(customTagFuncList) != 1 {
		panic("unexpected condition")
	}

	return customTagFuncList[0], nil
}

func getCustomTagOptionValues(commonCmdData *CmdData) []string {
	if commonCmdData.UseCustomTag != nil {
		if *commonCmdData.UseCustomTag != "" {
			return []string{*commonCmdData.UseCustomTag}
		}

		return nil
	} else {
		return getAddCustomTag(commonCmdData)
	}
}
