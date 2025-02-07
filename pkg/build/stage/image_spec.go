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
	"github.com/werf/werf/v2/pkg/dockerfile"
	"github.com/werf/werf/v2/pkg/dockerfile/frontend"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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
		newConfig.Volumes = modifyVolumes(ctx, imageInfo.Volumes, s.imageSpec.RemoveVolumes, s.imageSpec.Volumes)

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
		newConfig.ClearCmd = s.imageSpec.ClearCmd
		newConfig.ClearEntrypoint = s.imageSpec.ClearEntrypoint

		stageImage.Image.SetImageSpecConfig(&newConfig)
	}

	return nil
}

func (s *ImageSpecStage) GetDependencies(_ context.Context, _ Conveyor, _ container_backend.ContainerBackend, _, _ *StageImage, _ container_backend.BuildContextArchiver) (string, error) {
	var args []string

	args = append(args, "1")
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
	args = append(args, strings.Join(s.imageSpec.Entrypoint, " "))
	args = append(args, s.imageSpec.WorkingDir)
	args = append(args, s.imageSpec.StopSignal)
	args = append(args, fmt.Sprint(s.imageSpec.Healthcheck))

	return util.Sha256Hash(args...), nil
}

func modifyLabels(ctx context.Context, labels, addLabels map[string]string, removeLabels []string, clearWerfLabels bool) (map[string]string, error) {
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
			if !clearWerfLabels && strings.HasPrefix(key, "werf") {
				global_warnings.GlobalWarningLn(ctx, fmt.Sprintf("Removal of the werf label %q requires explicit use of the `clearWerfLabels` directive. Some labels are purely informational, while others are essential for cleanup operations.", key))
				continue
			}
			delete(labels, key)
			continue
		}

		if clearWerfLabels && strings.HasPrefix(key, "werf") {
			delete(labels, key)
		}
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

	for _, key := range removeKeys {
		delete(baseEnvMap, key)
	}

	envMapToExpand := make(map[string]string)
	for key, value := range baseEnvMap {
		envMapToExpand[key] = value
	}
	for key, value := range addKeysMap {
		// Do not add PATH to the baseEnvMap, because it is a special case for the expandEnv function.
		if key != "PATH" {
			baseEnvMap[key] = value
		}

		envMapToExpand[key] = value
	}

	expandedEnvMap, err := expandEnv(baseEnvMap, envMapToExpand, dockerfile.ExpandOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to expand env: %w", err)
	}

	resultEnv := make([]string, 0, len(baseEnvMap))
	for key, value := range expandedEnvMap {
		resultEnv = append(resultEnv, fmt.Sprintf("%s=%s", key, value))
	}

	return resultEnv, nil
}

func modifyVolumes(_ context.Context, volumes map[string]struct{}, removeVolumes, addVolumes []string) map[string]struct{} {
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

func expandEnv(baseEnvMap, envMapToExpand map[string]string, opts dockerfile.ExpandOptions) (map[string]string, error) {
	expander := frontend.NewShlexExpanderFactory('\\').GetExpander(opts)
	res := make(map[string]string)
	for k, v := range envMapToExpand {
		newValue, err := expander.ProcessWordWithMap(v, baseEnvMap)
		if err != nil {
			return nil, fmt.Errorf("error processing word %q: %w", v, err)
		}
		res[k] = newValue
	}
	return res, nil
}
