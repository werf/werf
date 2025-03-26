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

	"github.com/werf/common-go/pkg/util"
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
	werfLabelsGlobalWarning = `The "werf" and "werf-parent-stage-id" labels cannot be removed within the imageSpec stage, as they are essential for the proper operation of host and container registry cleanup.

If you need to remove all werf labels, use the werf export command. By default, this command removes all werf labels and fully detaches images from werf control, transferring host and container registry cleanup entirely to the user.

To disable this warning, explicitly set the "keepEssentialWerfLabels" directive.`
)

type ImageSpecStage struct {
	*BaseStage
	imageSpec *config.ImageSpec
}

func GenerateImageSpecStage(imageSpec *config.ImageSpec, baseStageOptions *BaseStageOptions) *ImageSpecStage {
	return newImageSpecStage(imageSpec, baseStageOptions)
}

func newImageSpecStage(imageSpec *config.ImageSpec, baseStageOptions *BaseStageOptions) *ImageSpecStage {
	s := &ImageSpecStage{}
	s.imageSpec = imageSpec
	s.BaseStage = NewBaseStage(ImageSpec, baseStageOptions)
	return s
}

func (s *ImageSpecStage) IsBuildable() bool {
	return false
}

func (s *ImageSpecStage) IsMutable() bool {
	return true
}

func (s *ImageSpecStage) PrepareImage(ctx context.Context, _ Conveyor, _ container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, _ container_backend.BuildContextArchiver) error {
	if s.imageSpec != nil {
		imageInfo := prevBuiltImage.Image.GetStageDesc().Info

		newConfig := image.Config{
			Author:     s.imageSpec.Author,
			User:       s.imageSpec.User,
			Entrypoint: s.imageSpec.Entrypoint,
			Cmd:        s.imageSpec.Cmd,
			WorkingDir: s.imageSpec.WorkingDir,
			StopSignal: s.imageSpec.StopSignal,
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

		resultLabels, err := s.modifyLabels(ctx, imageInfo.Labels, s.imageSpec.Labels, s.imageSpec.RemoveLabels, s.imageSpec.KeepEssentialWerfLabels)
		if err != nil {
			return fmt.Errorf("unable to modify labels: %s", err)
		}

		newConfig.Labels = resultLabels
		newConfig.Env, err = modifyEnv(imageInfo.Env, s.imageSpec.RemoveEnv, s.imageSpec.Env)
		if err != nil {
			return fmt.Errorf("unable to modify env: %w", err)
		}
		newConfig.Volumes = modifyVolumes(imageInfo.Volumes, s.imageSpec.RemoveVolumes, s.imageSpec.Volumes)

		if s.imageSpec.Expose != nil {
			newConfig.ExposedPorts = make(map[string]struct{}, len(s.imageSpec.Expose))
			for _, expose := range s.imageSpec.Expose {
				newConfig.ExposedPorts[expose] = struct{}{}
			}
		}

		if s.imageSpec.Healthcheck != nil {
			newConfig.HealthConfig = &image.HealthConfig{
				Test:        s.imageSpec.Healthcheck.Test,
				Interval:    toDuration(s.imageSpec.Healthcheck.Interval),
				Timeout:     toDuration(s.imageSpec.Healthcheck.Timeout),
				StartPeriod: toDuration(s.imageSpec.Healthcheck.StartPeriod),
				Retries:     s.imageSpec.Healthcheck.Retries,
			}
		}

		newConfig.ClearHistory = s.imageSpec.ClearHistory

		stageImage.Image.SetImageSpecConfig(&newConfig)
	}

	return nil
}

const imageSpecStageCacheVersion = "2"

func (s *ImageSpecStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, imageSpecStageCacheVersion)
	args = append(args, s.imageSpec.Author)
	args = append(args, fmt.Sprint(s.imageSpec.ClearHistory))

	args = append(args, fmt.Sprint(s.imageSpec.ClearWerfLabels))
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveLabels)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveVolumes)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveEnv)...)

	args = append(args, sortSliceWithNewSlice(s.imageSpec.Volumes)...)
	args = append(args, mapToSortedArgs(s.imageSpec.Labels)...)
	args = append(args, mapToSortedArgs(s.imageSpec.Env)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.Expose)...)
	args = append(args, s.imageSpec.User)
	args = append(args, strings.Join(s.imageSpec.Cmd, " "))
	args = append(args, fmt.Sprint(s.imageSpec.ClearCmd))
	args = append(args, strings.Join(s.imageSpec.Entrypoint, " "))
	args = append(args, fmt.Sprint(s.imageSpec.ClearEntrypoint))
	args = append(args, s.imageSpec.WorkingDir)
	args = append(args, s.imageSpec.StopSignal)
	args = append(args, fmt.Sprint(s.imageSpec.Healthcheck))

	return util.Sha256Hash(args...), nil
}

func (s *ImageSpecStage) modifyLabels(ctx context.Context, labels, addLabels map[string]string, removeLabels []string, keepEssentialWerfLabels bool) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}

	serviceLabels := s.stageImage.Image.GetBuildServiceLabels()
	labels = util.MergeMaps(labels, serviceLabels)

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

	exactMatches := make(map[string]struct{})
	var regexPatterns []*regexp.Regexp

	for _, pattern := range removeLabels {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			expr := fmt.Sprintf("^%s$", pattern[1:len(pattern)-1])
			re, err := regexp.Compile(expr)
			if err != nil {
				return nil, err
			}
			regexPatterns = append(regexPatterns, re)
		} else {
			exactMatches[pattern] = struct{}{}
		}
	}

	matchFunc := func(key string) bool {
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

	shouldPrintGlobalWarn := false
	for key := range labels {
		if !matchFunc(key) {
			continue
		}

		if key == image.WerfLabel || key == image.WerfParentStageID {
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

	for _, key := range removeKeys {
		delete(baseEnvMap, key)
	}

	// FIXME: (v3) This is a temporary solution to remove werf SSH_AUTH_SOCK that persist after build.
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
