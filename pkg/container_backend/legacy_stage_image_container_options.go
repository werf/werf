package container_backend

import (
	"context"
	"fmt"
	"sort"
)

type LegacyCommitChangeOptions struct {
	ExactValues bool
}

type LegacyStageImageContainerOptions struct {
	dockerServerVersion string

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
	Network     string
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

func (co *LegacyStageImageContainerOptions) AddNetwork(network string) {
	co.Network = network
}

func (co *LegacyStageImageContainerOptions) merge(co2 *LegacyStageImageContainerOptions) *LegacyStageImageContainerOptions {
	mergedCo := newLegacyStageContainerOptions()

	mergedCo.Volume = co.Volume
	mergedCo.Volume = append(mergedCo.Volume, co2.Volume...)

	mergedCo.VolumesFrom = co.VolumesFrom
	mergedCo.VolumesFrom = append(mergedCo.VolumesFrom, co2.VolumesFrom...)

	mergedCo.Expose = co.Expose
	mergedCo.Expose = append(mergedCo.Expose, co2.Expose...)

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

	if co2.Network == "" {
		mergedCo.Network = co.Network
	} else {
		mergedCo.Network = co2.Network
	}

	return mergedCo
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func sortStrings(arr []string) []string {
	sort.Strings(arr)
	return arr
}

func (co *LegacyStageImageContainerOptions) toRunArgs() ([]string, error) {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("--volume=%s", volume))
	}

	for _, volumesFrom := range co.VolumesFrom {
		args = append(args, fmt.Sprintf("--volumes-from=%s", volumesFrom))
	}

	for _, k := range sortStrings(getKeys(co.Env)) {
		args = append(args, fmt.Sprintf("--env=%s=%v", k, co.Env[k]))
	}

	for _, k := range sortStrings(getKeys(co.Label)) {
		args = append(args, fmt.Sprintf("--label=%s=%v", k, co.Label[k]))
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

	if co.Network != "" {
		args = append(args, fmt.Sprintf("--network=%s", co.Network))
	}

	return args, nil
}

func (co *LegacyStageImageContainerOptions) toCommitChanges(opts LegacyCommitChangeOptions) []string {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("VOLUME %s", escapeVolume(volume, opts.ExactValues)))
	}

	for _, expose := range co.Expose {
		args = append(args, fmt.Sprintf("EXPOSE %s", defaultEscape(expose, opts.ExactValues)))
	}

	for _, k := range sortStrings(getKeys(co.Env)) {
		args = append(args, fmt.Sprintf("ENV %s=%s", k, defaultEscape(co.Env[k], opts.ExactValues)))
	}

	for _, k := range sortStrings(getKeys(co.Label)) {
		args = append(args, fmt.Sprintf("LABEL %s=%s", k, defaultEscape(co.Label[k], opts.ExactValues)))
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

func (co *LegacyStageImageContainerOptions) prepareCommitChanges(ctx context.Context, opts LegacyCommitChangeOptions) ([]string, error) {
	var args []string

	for _, volume := range co.Volume {
		args = append(args, fmt.Sprintf("VOLUME %s", escapeVolume(volume, opts.ExactValues)))
	}

	for _, expose := range co.Expose {
		args = append(args, fmt.Sprintf("EXPOSE %s", defaultEscape(expose, opts.ExactValues)))
	}

	for _, k := range sortStrings(getKeys(co.Env)) {
		args = append(args, fmt.Sprintf("ENV %s=%s", k, defaultEscape(co.Env[k], opts.ExactValues)))
	}

	for _, k := range sortStrings(getKeys(co.Label)) {
		args = append(args, fmt.Sprintf("LABEL %s=%s", k, defaultEscape(co.Label[k], opts.ExactValues)))
	}

	if co.Workdir != "" {
		args = append(args, fmt.Sprintf("WORKDIR %s", co.Workdir))
	}

	if co.User != "" {
		args = append(args, fmt.Sprintf("USER %s", co.User))
	}

	var entrypoint string
	if co.Entrypoint != "" {
		entrypoint = co.Entrypoint
	} else {
		entrypoint = "[\"\"]"
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

// TODO(major): remove escaping
func escapeVolume(volume string, exactValues bool) string {
	if exactValues {
		return fmt.Sprintf("[%s]", quoteValue(volume))
	}
	return volume
}

func defaultEscape(value string, exactValues bool) string {
	return doEscape(value, exactValues, quoteValue)
}

func doEscape(value string, exactValues bool, escaper func(string) string) string {
	if exactValues {
		return escaper(value)
	}
	return value
}

func quoteValue(value string) string {
	return fmt.Sprintf("%q", value)
}
