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
)

var ErrImageSpecMutateStageIsNotSupported = fmt.Errorf("error running image spec mutate stage: can't mutate local image. please provide repo with --repo flag or WERF_REPO env variable")

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

func (s *ImageSpecStage) PrepareImage(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage, _ container_backend.BuildContextArchiver) error {
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

		newLabels, err := modifyLabels(imageInfo.Labels, s.imageSpec.Labels, s.imageSpec.RemoveLabels, s.imageSpec.RemoveWerfLabels)
		if err != nil {
			return fmt.Errorf("unable to modify labels: %s", err)
		}

		newConfig.Labels = newLabels
		newConfig.Env = modifyEnv(imageInfo.Env, s.imageSpec.RemoveEnv, s.imageSpec.Env)
		newConfig.Volumes = modifyVolumes(imageInfo.Volumes, s.imageSpec.RemoveVolumes, s.imageSpec.Volumes)

		for _, expose := range s.imageSpec.Expose {
			newConfig.ExposedPorts[expose] = struct{}{}
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
		newConfig.ClearCmd = s.imageSpec.ClearCmd
		newConfig.ClearEntrypoint = s.imageSpec.ClearEntrypoint

		stageImage.Image.SetImageSpecConfig(&newConfig)
	}

	return nil
}

func (s *ImageSpecStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, s.imageSpec.Author)
	args = append(args, fmt.Sprint(s.imageSpec.ClearHistory))

	args = append(args, fmt.Sprint(s.imageSpec.RemoveWerfLabels))
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveLabels)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveVolumes)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.RemoveEnv)...)

	args = append(args, sortSliceWithNewSlice(s.imageSpec.Volumes)...)
	args = append(args, mapToSortedArgs(s.imageSpec.Labels)...)
	args = append(args, mapToSortedArgs(s.imageSpec.Env)...)
	args = append(args, sortSliceWithNewSlice(s.imageSpec.Expose)...)
	args = append(args, s.imageSpec.User)
	args = append(args, strings.Join(s.imageSpec.Cmd, " "))
	args = append(args, strings.Join(s.imageSpec.Entrypoint, " "))
	args = append(args, s.imageSpec.WorkingDir)
	args = append(args, s.imageSpec.StopSignal)
	args = append(args, fmt.Sprint(s.imageSpec.Healthcheck))

	return util.Sha256Hash(args...), nil
}

func modifyLabels(labels, addLabels map[string]string, removeLabels []string, removeWerfLabels bool) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}

	shouldRemove := func(key string) (bool, error) {
		for _, pattern := range removeLabels {
			match, err := func() (bool, error) {
				if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
					expr := fmt.Sprintf("^%s$", pattern[1:len(pattern)-1])
					re, err := regexp.Compile(expr)
					if err != nil {
						return false, err
					}
					return re.MatchString(key), nil
				}
				return pattern == key, nil
			}()
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	}

	for key := range labels {
		match, err := shouldRemove(key)
		if err != nil {
			return nil, err
		}
		if match {
			delete(labels, key)
			continue
		}
		if removeWerfLabels && strings.HasPrefix(key, "werf") {
			delete(labels, key)
		}
	}

	for key, value := range addLabels {
		labels[key] = value
	}

	return labels, nil
}

func modifyEnv(env, removeKeys []string, addMap map[string]string) []string {
	envMap := make(map[string]string, len(env))
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for _, key := range removeKeys {
		delete(envMap, key)
	}

	for key, value := range addMap {
		envMap[key] = value
	}

	result := make([]string, 0, len(envMap))
	for key, value := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
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
	new := append([]string(nil), original...)
	sort.Strings(new)
	return new
}

func toDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
