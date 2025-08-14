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

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/build/signing"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/docker_registry/api"
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
	newConfig ImageSpecConfig
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
		imageInfo := prevBuiltImage.Image.GetStageDesc().Info
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
				newConfig.HealthConfig = &HealthConfig{
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

func (s *ImageSpecStage) MutateImage(ctx context.Context, registry docker_registry.Interface, prevBuiltImage, stageImage *StageImage, _ signing.ManifestSigningOptions) error {
	src := prevBuiltImage.Image.Name()
	dest := stageImage.Image.Name()

	return registry.MutateAndPushImage(ctx, src, dest, api.WithConfigFileMutation(func(ctx context.Context, config *v1.ConfigFile) (*v1.ConfigFile, error) {
		updateConfigFile(s.newConfig, config)
		return config, nil
	}))
}

func updateConfigFile(updates ImageSpecConfig, target *v1.ConfigFile) {
	if updates.Author != "" {
		target.Author = updates.Author
	}
	if updates.ClearHistory {
		target.History = []v1.History{}
	}
	if updates.Volumes != nil {
		target.Config.Volumes = updates.Volumes
	}
	if updates.Labels != nil {
		target.Config.Labels = updates.Labels
	}

	target.Config.Env = updates.Env

	if updates.ExposedPorts != nil {
		target.Config.ExposedPorts = updates.ExposedPorts
	}
	if updates.ClearUser {
		target.Config.User = ""
	}
	if updates.User != "" {
		target.Config.User = updates.User
	}
	if updates.ClearCmd {
		target.Config.Cmd = []string{}
	}
	if len(updates.Cmd) > 0 {
		target.Config.Cmd = updates.Cmd
	}
	if updates.ClearEntrypoint {
		target.Config.Entrypoint = []string{}
	}
	if len(updates.Entrypoint) > 0 {
		target.Config.Entrypoint = updates.Entrypoint
	}
	if updates.ClearWorkingDir {
		target.Config.WorkingDir = ""
	}
	if updates.WorkingDir != "" {
		target.Config.WorkingDir = updates.WorkingDir
	}
	if updates.StopSignal != "" {
		target.Config.StopSignal = updates.StopSignal
	}
	if updates.HealthConfig != nil {
		target.Config.Healthcheck = &v1.HealthConfig{
			Test:        updates.HealthConfig.Test,
			Interval:    updates.HealthConfig.Interval,
			Timeout:     updates.HealthConfig.Timeout,
			StartPeriod: updates.HealthConfig.StartPeriod,
			Retries:     updates.HealthConfig.Retries,
		}
	}
}

func (s *ImageSpecStage) baseConfig() ImageSpecConfig {
	newConfig := ImageSpecConfig{
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

// ImageSpecConfig represents OCI image configuration
// https://github.com/opencontainers/image-spec/blob/main/config.md
type ImageSpecConfig struct {
	Created         string              `json:"created"`
	Author          string              `json:"author"`
	User            string              `json:"User"`
	ExposedPorts    map[string]struct{} `json:"ExposedPorts"`
	Env             []string            `json:"Env"`
	Entrypoint      []string            `json:"Entrypoint"`
	Cmd             []string            `json:"Cmd"`
	Volumes         map[string]struct{} `json:"Volumes"`
	WorkingDir      string              `json:"WorkingDir"`
	Labels          map[string]string   `json:"Labels"`
	StopSignal      string              `json:"StopSignal"`
	HealthConfig    *HealthConfig       `json:"Healthcheck,omitempty"`
	ClearHistory    bool
	ClearCmd        bool
	ClearEntrypoint bool
	ClearUser       bool
	ClearWorkingDir bool
}

type HealthConfig struct {
	Test        []string      `json:",omitempty"`
	Interval    time.Duration `json:",omitempty"`
	Timeout     time.Duration `json:",omitempty"`
	StartPeriod time.Duration `json:",omitempty"`
	Retries     int           `json:",omitempty"`
}
