package image

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/flant/dapp/pkg/docker"
)

type StageContainerOptions struct {
	Volume      []string          `json:"volume"`
	VolumesFrom []string          `json:"volumes-from"`
	Expose      []string          `json:"expose"`
	Env         map[string]string `json:"env"`
	Label       map[string]string `json:"label"`
	Cmd         []string          `json:"cmd"`
	Onbuild     []string          `json:"onbuild"`
	Workdir     string            `json:"workdir"`
	User        string            `json:"user"`
	Entrypoint  []string          `json:"entrypoint"`
}

func newStageContainerOptions() *StageContainerOptions {
	c := &StageContainerOptions{}
	c.Env = make(map[string]string)
	c.Label = make(map[string]string)
	return c
}

func (co *StageContainerOptions) AddVolume(volumes ...string) {
	co.Volume = append(co.Volume, volumes...)
}

func (co *StageContainerOptions) AddVolumeFrom(volumesFrom ...string) {
	co.VolumesFrom = append(co.VolumesFrom, volumesFrom...)
}

func (co *StageContainerOptions) AddExpose(exposes ...string) {
	co.Expose = append(co.Expose, exposes...)
}

func (co *StageContainerOptions) AddEnv(envs map[string]string) {
	for env, value := range envs {
		co.Env[env] = value
	}
}

func (co *StageContainerOptions) AddLabel(labels map[string]string) {
	for label, value := range labels {
		co.Label[label] = value
	}
}

func (co *StageContainerOptions) AddCmd(cmds ...string) {
	co.Cmd = append(co.Cmd, cmds...)
}

func (co *StageContainerOptions) AddOnbuild(onbuilds ...string) {
	co.Onbuild = append(co.Onbuild, onbuilds...)
}

func (co *StageContainerOptions) AddWorkdir(workdir string) {
	co.Workdir = workdir
}

func (co *StageContainerOptions) AddUser(user string) {
	co.User = user
}

func (co *StageContainerOptions) AddEntrypoint(entrypoints ...string) {
	co.Entrypoint = append(co.Entrypoint, entrypoints...)
}

func (co *StageContainerOptions) merge(co2 *StageContainerOptions) *StageContainerOptions {
	mergedCo := newStageContainerOptions()
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

	if len(co2.Onbuild) == 0 {
		mergedCo.Onbuild = co.Onbuild
	} else {
		mergedCo.Onbuild = co2.Onbuild
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

	return mergedCo
}

func (co *StageContainerOptions) toRunArgs() ([]string, error) {
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

	if len(co.Entrypoint) == 1 {
		args = append(args, fmt.Sprintf("--entrypoint=%s", co.Entrypoint[0]))
	} else if len(co.Entrypoint) != 0 {
		return nil, fmt.Errorf("`Entrypoint` value `%v` isn't supported in run command (only string)", co.Entrypoint)
	}

	return args, nil
}

func (co *StageContainerOptions) toCommitChanges() ([]string, error) {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("Volume %s", volume))
	}

	for _, expose := range co.Expose {
		args = append(args, fmt.Sprintf("Expose %s", expose))
	}

	for key, value := range co.Env {
		args = append(args, fmt.Sprintf("ENV %s=%v", key, value))
	}

	for key, value := range co.Label {
		args = append(args, fmt.Sprintf("Label %s=%v", key, value))
	}

	if len(co.Cmd) == 0 {
		cmd, err := getEmptyCmdOrEntrypointInstructionValue()
		if err != nil {
			return nil, fmt.Errorf("container options preparing failed: %s", err.Error())
		}
		args = append(args, fmt.Sprintf("Cmd %s", cmd))
	} else if len(co.Cmd) != 0 {
		args = append(args, fmt.Sprintf("Cmd [\"%s\"]", strings.Join(co.Cmd, "\", \"")))
	}

	if len(co.Onbuild) != 0 {
		args = append(args, fmt.Sprintf("Onbuild %s", strings.Join(co.Onbuild, " ")))
	}

	if co.Workdir != "" {
		args = append(args, fmt.Sprintf("Workdir %s", co.Workdir))
	}

	if co.User != "" {
		args = append(args, fmt.Sprintf("User %s", co.User))
	}

	if len(co.Entrypoint) == 0 {
		entrypoint, err := getEmptyCmdOrEntrypointInstructionValue()
		if err != nil {
			return nil, fmt.Errorf("container options preparing failed: %s", err.Error())
		}
		args = append(args, fmt.Sprintf("Entrypoint %s", entrypoint))
	} else if len(co.Entrypoint) != 0 {
		args = append(args, fmt.Sprintf("Entrypoint [\"%s\"]", strings.Join(co.Entrypoint, "\", \"")))
	}

	return args, nil
}

func getEmptyCmdOrEntrypointInstructionValue() (string, error) {
	v, err := docker.ServerVersion()
	if err != nil {
		return "", err
	}

	serverVersion, err := version.NewVersion(v.Version)
	if err != nil {
		return "", err
	}

	verifiableVersion, err := version.NewVersion("17.10")
	if err != nil {
		return "", err
	}

	if serverVersion.LessThan(verifiableVersion) {
		return "[]", nil
	} else {
		return "[\"\"]", nil
	}
}
