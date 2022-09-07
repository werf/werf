package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/container_backend"
	"github.com/werf/werf/pkg/util"
)

func GenerateDockerInstructionsStage(imageConfig *config.StapelImage, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	if imageConfig.Docker != nil {
		return newDockerInstructionsStage(imageConfig.Docker, baseStageOptions)
	}

	return nil
}

func newDockerInstructionsStage(instructions *config.Docker, baseStageOptions *NewBaseStageOptions) *DockerInstructionsStage {
	s := &DockerInstructionsStage{}
	s.instructions = instructions
	s.BaseStage = newBaseStage(DockerInstructions, baseStageOptions)
	return s
}

type DockerInstructionsStage struct {
	*BaseStage

	instructions *config.Docker
}

func (s *DockerInstructionsStage) GetDependencies(_ context.Context, c Conveyor, backend container_backend.ContainerBackend, _, _ *StageImage) (string, error) {
	var args []string

	if c.UseLegacyStapelBuilder(backend) && s.instructions.ExactValues {
		args = append(args, "exact-values:::")
	}

	args = append(args, s.instructions.Volume...)
	args = append(args, s.instructions.Expose...)
	args = append(args, mapToSortedArgs(s.instructions.Env)...)
	args = append(args, mapToSortedArgs(s.instructions.Label)...)
	args = append(args, s.instructions.Cmd)
	args = append(args, s.instructions.Entrypoint)
	args = append(args, s.instructions.Workdir)
	args = append(args, s.instructions.User)
	args = append(args, s.instructions.HealthCheck)

	return util.Sha256Hash(args...), nil
}

func mapToSortedArgs(h map[string]string) (result []string) {
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		result = append(result, key, h[key])
	}

	return
}

func (s *DockerInstructionsStage) PrepareImage(ctx context.Context, c Conveyor, cr container_backend.ContainerBackend, prevBuiltImage, stageImage *StageImage) error {
	if c.UseLegacyStapelBuilder(cr) {
		stageImage.Image.SetCommitChangeOptions(container_backend.LegacyCommitChangeOptions{ExactValues: s.instructions.ExactValues})
	}

	if err := s.BaseStage.PrepareImage(ctx, c, cr, prevBuiltImage, stageImage); err != nil {
		return err
	}

	if c.UseLegacyStapelBuilder(cr) {
		imageCommitChangeOptions := stageImage.Builder.LegacyStapelStageBuilder().Container().CommitChangeOptions()
		imageCommitChangeOptions.AddVolume(s.instructions.Volume...)
		imageCommitChangeOptions.AddExpose(s.instructions.Expose...)
		imageCommitChangeOptions.AddEnv(s.instructions.Env)
		imageCommitChangeOptions.AddLabel(s.instructions.Label)
		imageCommitChangeOptions.AddCmd(s.instructions.Cmd)
		imageCommitChangeOptions.AddEntrypoint(s.instructions.Entrypoint)
		imageCommitChangeOptions.AddUser(s.instructions.User)
		imageCommitChangeOptions.AddWorkdir(s.instructions.Workdir)
		imageCommitChangeOptions.AddHealthCheck(s.instructions.HealthCheck)
	} else {
		builder := stageImage.Builder.StapelStageBuilder()

		builder.AddVolumes(s.instructions.Volume).
			AddExpose(s.instructions.Expose).
			AddEnvs(s.instructions.Env).
			AddLabels(s.instructions.Label).
			SetUser(s.instructions.User).
			SetWorkdir(s.instructions.Workdir).
			SetHealthcheck(s.instructions.HealthCheck)

		if ep, err := CmdOrEntrypointStringToSlice(s.instructions.Entrypoint); err != nil {
			return fmt.Errorf("error converting ENTRYPOINT from string to slice: %w", err)
		} else {
			builder.SetEntrypoint(ep)
		}

		if cmd, err := CmdOrEntrypointStringToSlice(s.instructions.Cmd); err != nil {
			return fmt.Errorf("error converting CMD from string to slice: %w", err)
		} else {
			builder.SetCmd(cmd)
		}
	}

	return nil
}

func CmdOrEntrypointStringToSlice(cmdOrEntrypoint string) ([]string, error) {
	var result []string
	if len(cmdOrEntrypoint) > 0 {
		if string(cmdOrEntrypoint[0]) == "[" && string(cmdOrEntrypoint[len(cmdOrEntrypoint)-1]) == "]" {
			if err := json.Unmarshal([]byte(cmdOrEntrypoint), &result); err != nil {
				return nil, fmt.Errorf("error parsing to the JSON array: %w", err)
			}
		} else {
			result = []string{"/bin/sh", "-c", cmdOrEntrypoint}
		}
	}

	return result, nil
}
