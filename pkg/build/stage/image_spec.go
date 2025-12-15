package stage

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"regexp"
	"sort"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/ssh_agent"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

const (
	labelTemplateImage      = "image"
	labelTemplateProject    = "project"
	labelTemplateDelimiter  = "%"
	werfLabelsGlobalWarning = `The "werf", "werf-stage-content-digest" and "werf.io/parent-stage-id" labels cannot be removed within the imageSpec stage, as they are essential for the proper operation of host and container registry cleanup.

If you need to remove all werf labels, use the werf export command. By default, this command removes all werf labels and fully detaches images from werf control, transferring host and container registry cleanup entirely to the user.

To disable this warning, explicitly set the "keepEssentialWerfLabels" directive.`
)

type ImageSpecStage struct {
	*BaseStage
	imageSpec *config.ImageSpec
	newConfig image.SpecConfig
}

func GenerateImageSpecStage(imageSpec *config.ImageSpec, baseStageOptions *BaseStageOptions) *ImageSpecStage {
	return newImageSpecStage(imageSpec, baseStageOptions)
}

func newImageSpecStage(imageSpec *config.ImageSpec, baseStageOptions *BaseStageOptions) *ImageSpecStage {
	return &ImageSpecStage{
		imageSpec: imageSpec,
		BaseStage: NewBaseStage(ImageSpec, baseStageOptions),
	}
}

func (s *ImageSpecStage) IsBuildable() bool {
	return false
}

func (s *ImageSpecStage) IsMutable() bool {
	return true
}

func (s *ImageSpecStage) PrepareImage(ctx context.Context, _ Conveyor, _ container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, _ container_backend.BuildContextArchiver) error {
	if s.imageSpec != nil {
		// NOTE. We need a copy, because we mutate labels, volumes and envs.
		imageInfo := prevBuiltImage.Image.GetStageDesc().Info.GetCopy()

		if err := logboek.Context(ctx).Debug().LogBlock("-- ImageSpecStage.PrepareImage source image info").DoError(func() error {
			data, err := yaml.Marshal(imageInfo)
			if err != nil {
				return fmt.Errorf("unable to yaml marshal: %w", err)
			}

			logboek.Context(ctx).Debug().LogF(string(data))
			return nil
		}); err != nil {
			return err
		}

		newConfig := s.baseConfig()

		{
			// labels
			resultLabels, err := s.modifyLabels(ctx, imageInfo.Labels, s.imageSpec.Labels, s.imageSpec.RemoveLabels, s.imageSpec.KeepEssentialWerfLabels)
			if err != nil {
				return fmt.Errorf("unable to modify labels: %s", err)
			}
			newConfig.Labels = resultLabels
		}

		{
			// envs
			resultEnvs, err := modifyEnv(imageInfo.Env, s.imageSpec.RemoveEnv, s.imageSpec.Env)
			if err != nil {
				return fmt.Errorf("unable to modify env: %w", err)
			}
			newConfig.Env = resultEnvs
		}

		{
			// volumes
			newConfig.Volumes = modifyVolumes(imageInfo.Volumes, s.imageSpec.RemoveVolumes, s.imageSpec.Volumes)
		}

		{
			// expose
			if s.imageSpec.Expose != nil {
				newConfig.ExposedPorts = make(map[string]struct{}, len(s.imageSpec.Expose))
				for _, expose := range s.imageSpec.Expose {
					newConfig.ExposedPorts[expose] = struct{}{}
				}
			}
		}

		{
			// healthcheck
			if s.imageSpec.Healthcheck != nil {
				newConfig.HealthConfig = &image.HealthConfig{
					Test:        s.imageSpec.Healthcheck.Test,
					Interval:    toDuration(s.imageSpec.Healthcheck.Interval),
					Timeout:     toDuration(s.imageSpec.Healthcheck.Timeout),
					StartPeriod: toDuration(s.imageSpec.Healthcheck.StartPeriod),
					Retries:     s.imageSpec.Healthcheck.Retries,
				}
			}
		}

		// set config
		s.newConfig = newConfig

		if err := logboek.Context(ctx).Debug().LogBlock("-- ImageSpecStage.PrepareImage prepared image spec config").DoError(func() error {
			data, err := yaml.Marshal(s.newConfig)
			if err != nil {
				return fmt.Errorf("unable to yaml marshal: %w", err)
			}

			logboek.Context(ctx).Debug().LogF(string(data))
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *ImageSpecStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	// imageSpec
	args = append(args, s.imageSpec.Author)
	args = append(args, fmt.Sprint(s.imageSpec.ClearHistory))

	// imageSpec.config
	args = append(args, strings.Join(s.imageSpec.Cmd, " "))
	args = append(args, strings.Join(s.imageSpec.Entrypoint, " "))
	args = append(args, mapToSortedArgs(s.imageSpec.Env)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.Expose)...)
	args = append(args, fmt.Sprint(s.imageSpec.Healthcheck))
	args = append(args, mapToSortedArgs(s.imageSpec.Labels)...)
	args = append(args, s.imageSpec.StopSignal)
	args = append(args, s.imageSpec.User)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.Volumes)...)
	args = append(args, s.imageSpec.WorkingDir)

	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveLabels)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveVolumes)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveEnv)...)
	args = append(args, fmt.Sprint(s.imageSpec.KeepEssentialWerfLabels))

	args = append(args, fmt.Sprint(s.imageSpec.ClearCmd))
	args = append(args, fmt.Sprint(s.imageSpec.ClearEntrypoint))
	args = append(args, fmt.Sprint(s.imageSpec.ClearUser))
	args = append(args, fmt.Sprint(s.imageSpec.ClearWorkingDir))

	return util.Sha256Hash(args...), nil
}

type ImageMutatorPusher interface {
	MutateAndPushImage(ctx context.Context, src, dest string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error
}

func (s *ImageSpecStage) MutateImage(ctx context.Context, storage ImageMutatorPusher, prevBuiltImage, stageImage *StageImage) error {
	src := prevBuiltImage.Image.Name()
	dest := stageImage.Image.Name()
	return storage.MutateAndPushImage(ctx, src, dest, s.newConfig, stageImage.Image)
}

func (s *ImageSpecStage) baseConfig() image.SpecConfig {
	newConfig := image.SpecConfig{
		Author:          s.imageSpec.Author,
		User:            s.imageSpec.User,
		Entrypoint:      s.imageSpec.Entrypoint,
		Cmd:             s.imageSpec.Cmd,
		WorkingDir:      s.imageSpec.WorkingDir,
		StopSignal:      s.imageSpec.StopSignal,
		ClearHistory:    s.imageSpec.ClearHistory,
		ClearUser:       s.imageSpec.ClearUser,
		ClearWorkingDir: s.imageSpec.ClearWorkingDir,
	}

	// Entrypoint and Cmd handling.
	{
		// If CMD is defined from the base image, setting ENTRYPOINT will reset CMD to an empty value.
		// In this scenario, CMD must be defined in the current image to have a value.
		// rel https://docs.docker.com/reference/dockerfile/#understand-how-cmd-and-entrypoint-interact
		if s.imageSpec.Entrypoint != nil {
			newConfig.Entrypoint = s.imageSpec.Entrypoint
			if s.imageSpec.Cmd == nil {
				newConfig.ClearCmd = true
			}
		}

		if s.imageSpec.Cmd != nil {
			newConfig.Cmd = s.imageSpec.Cmd
		}

		newConfig.ClearCmd = newConfig.ClearCmd || s.imageSpec.ClearCmd
		newConfig.ClearEntrypoint = newConfig.ClearEntrypoint || s.imageSpec.ClearEntrypoint
	}
	return newConfig
}

func (s *ImageSpecStage) modifyLabels(ctx context.Context, labels, addLabels map[string]string, removeLabels []string, keepEssentialWerfLabels bool) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}

	serviceLabels := s.stageImage.Image.GetBuildServiceLabels()
	labels = util.MergeMaps(labels, serviceLabels)

	exactMatches, regexPatterns, err := compileRemovePatterns(removeLabels)
	if err != nil {
		return nil, err
	}

	shouldPrintGlobalWarn := false
	for key := range labels {
		if !matchKey(key, exactMatches, regexPatterns) {
			continue
		}

		if key == image.WerfLabel || key == image.WerfParentStageID || key == image.WerfStageContentDigestLabel {
			if !keepEssentialWerfLabels {
				shouldPrintGlobalWarn = true
			}
			continue
		}

		delete(labels, key)
	}

	if shouldPrintGlobalWarn {
		global_warnings.GlobalWarningLn(ctx, werfLabelsGlobalWarning)
	}

	processedAddLabels := make(map[string]string, len(addLabels))
	data := labelsTemplateData{
		Project: s.projectName,
		Image:   s.imageName,
	}

	for key, value := range addLabels {
		newKey, newValue, err := replaceLabelTemplate(key, value, data)
		if err != nil {
			return nil, err
		}
		processedAddLabels[newKey] = newValue
	}

	labels = util.MergeMaps(labels, processedAddLabels)

	return labels, nil
}

type labelsTemplateData struct {
	Project string
	Image   string
}

func replaceLabelTemplate(k, v string, data labelsTemplateData) (string, string, error) {
	funcMap := template.FuncMap{
		labelTemplateProject: func() string { return data.Project },
		labelTemplateImage:   func() string { return data.Image },
	}

	keyTmpl := template.New("key").Delims(labelTemplateDelimiter, labelTemplateDelimiter).Funcs(funcMap)
	valueTmpl := template.New("value").Delims(labelTemplateDelimiter, labelTemplateDelimiter).Funcs(funcMap)

	parsedKeyTmpl, err := keyTmpl.Parse(k)
	if err != nil {
		return "", "", err
	}

	parsedValueTmpl, err := valueTmpl.Parse(v)
	if err != nil {
		return "", "", err
	}

	var keyBuf, valueBuf bytes.Buffer

	if err := parsedKeyTmpl.Execute(&keyBuf, data); err != nil {
		return "", "", err
	}

	if err := parsedValueTmpl.Execute(&valueBuf, data); err != nil {
		return "", "", err
	}
	return keyBuf.String(), valueBuf.String(), nil
}

func modifyEnv(env, removeKeys []string, addKeysMap map[string]string) ([]string, error) {
	baseEnvMap := make(map[string]string, len(env))

	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			baseEnvMap[parts[0]] = parts[1]
		}
	}

	// FIXME: (v3) This is a temporary solution to remove werf commit envs that persist after build.
	removeKeys = append(removeKeys, []string{
		"WERF_COMMIT_HASH",
		"WERF_COMMIT_TIME_HUMAN",
		"WERF_COMMIT_TIME_UNIX",
	}...)

	exactMatches, regexPatterns, err := compileRemovePatterns(removeKeys)
	if err != nil {
		return nil, err
	}

	for key := range baseEnvMap {
		if matchKey(key, exactMatches, regexPatterns) {
			delete(baseEnvMap, key)
		}
	}

	if envValue, hasEnv := baseEnvMap[ssh_agent.SSHAuthSockEnv]; hasEnv && envValue == container_backend.SSHContainerAuthSockPath {
		delete(baseEnvMap, ssh_agent.SSHAuthSockEnv)
	}

	for k, v := range addKeysMap {
		newVal, err := shlexProcessWord(v, env)
		if err != nil {
			return nil, err
		}
		baseEnvMap[k] = newVal
	}

	newEnv := make([]string, 0, len(baseEnvMap))
	for k, v := range baseEnvMap {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return newEnv, nil
}

func modifyVolumes(volumes map[string]struct{}, removeVolumes, addVolumes []string) map[string]struct{} {
	if volumes == nil {
		volumes = make(map[string]struct{})
	}

	for _, volume := range removeVolumes {
		delete(volumes, volume)
	}

	for _, volume := range addVolumes {
		volumes[volume] = struct{}{}
	}

	return volumes
}

func sortSliceWithNewSlice(original []string) []string {
	result := append([]string(nil), original...)
	sort.Strings(result)
	return result
}

func toDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

func compileRemovePatterns(removePatterns []string) (map[string]struct{}, []*regexp.Regexp, error) {
	exactMatches := make(map[string]struct{})
	var regexPatterns []*regexp.Regexp

	for _, pattern := range removePatterns {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			expr := fmt.Sprintf("^%s$", pattern[1:len(pattern)-1])
			re, err := regexp.Compile(expr)
			if err != nil {
				return nil, nil, err
			}
			regexPatterns = append(regexPatterns, re)
		} else {
			exactMatches[pattern] = struct{}{}
		}
	}

	return exactMatches, regexPatterns, nil
}

func matchKey(key string, exactMatches map[string]struct{}, regexPatterns []*regexp.Regexp) bool {
	if _, found := exactMatches[key]; found {
		return true
	}
	for _, re := range regexPatterns {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}
