package stage

import (
	"context"
	"fmt"
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
	werfLabelsGlobalWarning      = "Removal of the werf labels requires explicit use of the clearWerfLabels directive. Some labels are purely informational, while others are essential for cleanup operations."
	werfLabelsHostCleanupWarning = "Removal of the %s label will affect host auto cleanup. Proper work of auto cleanup is not guaranteed."
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

		serviceLabels := s.stageImage.Image.GetBuildServiceLabels()
		mergedLabels := util.MergeMaps(s.imageSpec.Labels, serviceLabels)
		resultLabels, err := modifyLabels(ctx, mergedLabels, s.imageSpec.Labels, s.imageSpec.RemoveLabels, s.imageSpec.ClearWerfLabels)
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

func modifyLabels(ctx context.Context, labels, addLabels map[string]string, removeLabels []string, clearWerfLabels bool) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}

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

	shouldRemove := func(key string) bool {
		if _, found := exactMatches[key]; found {
			return true
		}
		for _, re := range regexPatterns {
			if re.MatchString(key) {
				return true
			}
		}
		if clearWerfLabels && strings.HasPrefix(key, "werf") {
			return true
		}
		return false
	}

	shouldPrintGlobalWarn := false
	var cleanupWarnKeys []string
	for key := range labels {
		if shouldRemove(key) {
			if !clearWerfLabels && strings.HasPrefix(key, "werf") {
				shouldPrintGlobalWarn = true
				if key == image.WerfStageDigestLabel {
					cleanupWarnKeys = append(cleanupWarnKeys, key)
				}
			} else {
				delete(labels, key)
			}
		}
	}

	if shouldPrintGlobalWarn {
		global_warnings.GlobalWarningLn(ctx, werfLabelsGlobalWarning)
	}

	if len(cleanupWarnKeys) > 0 {
		global_warnings.GlobalWarningLn(ctx, fmt.Sprintf(werfLabelsHostCleanupWarning, strings.Join(cleanupWarnKeys, "','")))
	}

	for key, value := range addLabels {
		labels[key] = value
	}

	return labels, nil
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
