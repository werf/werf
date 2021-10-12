package container_runtime

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"

	"github.com/werf/werf/pkg/docker"
)

type LegacyStageImageContainerOptions struct {
	Volume      []string
	VolumesFrom []string
	Expose      []string
	Env         map[string]string
	Label       map[string]string
	Cmd         string
	Workdir     string
	User        string
	Entrypoint  string
	HealthCheck string
}

func newLegacyStageContainerOptions() *LegacyStageImageContainerOptions {
	c := &LegacyStageImageContainerOptions{}
	c.Env = make(map[string]string)
	c.Label = make(map[string]string)
	return c
}

func (co *LegacyStageImageContainerOptions) AddVolume(volumes ...string) {
	co.Volume = append(co.Volume, volumes...)
}

func (co *LegacyStageImageContainerOptions) AddVolumeFrom(volumesFrom ...string) {
	co.VolumesFrom = append(co.VolumesFrom, volumesFrom...)
}

func (co *LegacyStageImageContainerOptions) AddExpose(exposes ...string) {
	co.Expose = append(co.Expose, exposes...)
}

func (co *LegacyStageImageContainerOptions) AddEnv(envs map[string]string) {
	for env, value := range envs {
		co.Env[env] = value
	}
}

func (co *LegacyStageImageContainerOptions) AddLabel(labels map[string]string) {
	for label, value := range labels {
		co.Label[label] = value
	}
}

func (co *LegacyStageImageContainerOptions) AddCmd(cmd string) {
	co.Cmd = cmd
}

func (co *LegacyStageImageContainerOptions) AddWorkdir(workdir string) {
	co.Workdir = workdir
}

func (co *LegacyStageImageContainerOptions) AddUser(user string) {
	co.User = user
}

func (co *LegacyStageImageContainerOptions) AddHealthCheck(check string) {
	co.HealthCheck = check
}

func (co *LegacyStageImageContainerOptions) AddEntrypoint(entrypoint string) {
	co.Entrypoint = entrypoint
}

func (co *LegacyStageImageContainerOptions) merge(co2 *LegacyStageImageContainerOptions) *LegacyStageImageContainerOptions {
	mergedCo := newLegacyStageContainerOptions()
	mergedCo.Volume = append(co.Volume, co2.Volume...)
	mergedCo.VolumesFrom = append(co.VolumesFrom, co2.VolumesFrom...)
	mergedCo.Expose = append(co.Expose, co2.Expose...)

	for env, value := range co.Env {
		mergedCo.Env[env] = value
	}
	for env, value := range co2.Env {
		mergedCo.Env[env] = value
	}

	for label, value := range co.Label {
		mergedCo.Label[label] = value
	}
	for label, value := range co2.Label {
		mergedCo.Label[label] = value
	}

	if len(co2.Cmd) == 0 {
		mergedCo.Cmd = co.Cmd
	} else {
		mergedCo.Cmd = co2.Cmd
	}

	if co2.Workdir == "" {
		mergedCo.Workdir = co.Workdir
	} else {
		mergedCo.Workdir = co2.Workdir
	}

	if co2.User == "" {
		mergedCo.User = co.User
	} else {
		mergedCo.User = co2.User
	}

	if len(co2.Entrypoint) == 0 {
		mergedCo.Entrypoint = co.Entrypoint
	} else {
		mergedCo.Entrypoint = co2.Entrypoint
	}

	if co2.HealthCheck == "" {
		mergedCo.HealthCheck = co.HealthCheck
	} else {
		mergedCo.HealthCheck = co2.HealthCheck
	}

	return mergedCo
}

func (co *LegacyStageImageContainerOptions) toRunArgs() ([]string, error) {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("--volume=%s", volume))
	}

	for _, volumesFrom := range co.VolumesFrom {
		args = append(args, fmt.Sprintf("--volumes-from=%s", volumesFrom))
	}

	for key, value := range co.Env {
		args = append(args, fmt.Sprintf("--env=%s=%v", key, value))
	}

	for key, value := range co.Label {
		args = append(args, fmt.Sprintf("--label=%s=%v", key, value))
	}

	if co.User != "" {
		args = append(args, fmt.Sprintf("--user=%s", co.User))
	}

	if co.Workdir != "" {
		args = append(args, fmt.Sprintf("--workdir=%s", co.Workdir))
	}

	if co.Entrypoint != "" {
		args = append(args, fmt.Sprintf("--entrypoint=%s", co.Entrypoint))
	}

	return args, nil
}

func (co *LegacyStageImageContainerOptions) toCommitChanges() []string {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("VOLUME %s", volume))
	}

	for _, expose := range co.Expose {
		args = append(args, fmt.Sprintf("EXPOSE %s", expose))
	}

	for key, value := range co.Env {
		args = append(args, fmt.Sprintf("ENV %s=%v", key, value))
	}

	for key, value := range co.Label {
		args = append(args, fmt.Sprintf("LABEL %s=%v", key, value))
	}

	if co.Cmd != "" {
		args = append(args, fmt.Sprintf("CMD %s", co.Cmd))
	}

	if co.Workdir != "" {
		args = append(args, fmt.Sprintf("WORKDIR %s", co.Workdir))
	}

	if co.User != "" {
		args = append(args, fmt.Sprintf("USER %s", co.User))
	}

	if co.Entrypoint != "" {
		args = append(args, fmt.Sprintf("ENTRYPOINT %s", co.Entrypoint))
	}

	if co.HealthCheck != "" {
		args = append(args, fmt.Sprintf("HEALTHCHECK %s", co.HealthCheck))
	}

	return args
}

func (co *LegacyStageImageContainerOptions) prepareCommitChanges(ctx context.Context) ([]string, error) {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("VOLUME %s", volume))
	}

	for _, expose := range co.Expose {
		args = append(args, fmt.Sprintf("EXPOSE %s", expose))
	}

	for key, value := range co.Env {
		args = append(args, fmt.Sprintf("ENV %s=%v", key, value))
	}

	for key, value := range co.Label {
		args = append(args, fmt.Sprintf("LABEL %s=%v", key, value))
	}

	if co.Workdir != "" {
		args = append(args, fmt.Sprintf("WORKDIR %s", co.Workdir))
	}

	if co.User != "" {
		args = append(args, fmt.Sprintf("USER %s", co.User))
	}

	var entrypoint string
	var err error
	if co.Entrypoint != "" {
		entrypoint = co.Entrypoint
	} else {
		entrypoint, err = getEmptyEntrypointInstructionValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("container options preparing failed: %s", err.Error())
		}
	}

	args = append(args, fmt.Sprintf("ENTRYPOINT %s", entrypoint))

	if co.Cmd != "" {
		args = append(args, fmt.Sprintf("CMD %s", co.Cmd))
	} else if co.Entrypoint == "" {
		args = append(args, "CMD []")
	}

	if co.HealthCheck != "" {
		args = append(args, fmt.Sprintf("HEALTHCHECK %s", co.HealthCheck))
	}

	return args, nil
}

func getEmptyEntrypointInstructionValue(ctx context.Context) (string, error) {
	v, err := docker.ServerVersion(ctx)
	if err != nil {
		return "", err
	}

	serverVersion, err := version.NewVersion(v.Version)
	if err != nil {
		return "", err
	}

	serverVersionMajor := serverVersion.Segments()[0]
	if serverVersionMajor >= 17 {
		serverVersionMinor := serverVersion.Segments()[1]
		isOldValueFormat := serverVersionMajor == 17 && serverVersionMinor < 10
		if isOldValueFormat {
			return "[]", nil
		}
	}

	return "[\"\"]", nil
}
