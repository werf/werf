package frontend

import (
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/dockerfile"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

func DockerfileStageFromBuildkitStage(d *dockerfile.Dockerfile, stage instructions.Stage) (*dockerfile.DockerfileStage, error) {
	var i []dockerfile.InstructionInterface

	for _, cmd := range stage.Commands {
		switch typedCmd := cmd.(type) {
		case *instructions.AddCommand:
			src, dst := extractSrcAndDst(typedCmd.SourcesAndDest)
			i = append(i, dockerfile_instruction.NewAdd(src, dst, typedCmd.Chown, typedCmd.Chmod))
		case *instructions.ArgCommand:
			i = append(i, dockerfile_instruction.NewArg(typedCmd.Args))
		case *instructions.CmdCommand:
			i = append(i, dockerfile_instruction.NewCmd(typedCmd.CmdLine, typedCmd.PrependShell))
		case *instructions.CopyCommand:
			src, dst := extractSrcAndDst(typedCmd.SourcesAndDest)
			i = append(i, dockerfile_instruction.NewCopy(typedCmd.From, src, dst, typedCmd.Chown, typedCmd.Chmod))
		case *instructions.EntrypointCommand:
			i = append(i, dockerfile_instruction.NewEntrypoint(typedCmd.CmdLine, typedCmd.PrependShell))
		case *instructions.EnvCommand:
			i = append(i, dockerfile_instruction.NewEnv(extractKeyValuePairsAsMap(typedCmd.Env)))
		case *instructions.ExposeCommand:
			i = append(i, dockerfile_instruction.NewExpose(typedCmd.Ports))
		case *instructions.HealthCheckCommand:
			i = append(i, dockerfile_instruction.NewHealthcheck(typedCmd.Health))
		case *instructions.LabelCommand:
			i = append(i, dockerfile_instruction.NewLabel(extractKeyValuePairsAsMap(typedCmd.Labels)))
		case *instructions.MaintainerCommand:
			i = append(i, dockerfile_instruction.NewMaintainer(typedCmd.Maintainer))
		case *instructions.OnbuildCommand:
			i = append(i, dockerfile_instruction.NewOnBuild(typedCmd.Expression))
		case *instructions.RunCommand:
			network := dockerfile_instruction.NewNetworkType(instructions.GetNetwork(typedCmd))
			security := dockerfile_instruction.NewSecurityType(instructions.GetSecurity(typedCmd))
			mounts := instructions.GetMounts(typedCmd)
			i = append(i, dockerfile_instruction.NewRun(typedCmd.CmdLine, typedCmd.PrependShell, mounts, network, security))
		case *instructions.ShellCommand:
			i = append(i, dockerfile_instruction.NewShell(typedCmd.Shell))
		case *instructions.StopSignalCommand:
			i = append(i, dockerfile_instruction.NewStopSignal(typedCmd.Signal))
		case *instructions.UserCommand:
			i = append(i, dockerfile_instruction.NewUser(typedCmd.User))
		case *instructions.VolumeCommand:
			i = append(i, dockerfile_instruction.NewVolume(typedCmd.Volumes))
		case *instructions.WorkdirCommand:
			i = append(i, dockerfile_instruction.NewWorkdir(typedCmd.Path))
		}
	}

	return dockerfile.NewDockerfileStage(d, i), nil
}

func extractSrcAndDst(sourcesAndDest instructions.SourcesAndDest) ([]string, string) {
	if len(sourcesAndDest) < 2 {
		panic(fmt.Sprintf("unexpected buildkit instruction source and destination: %#v", sourcesAndDest))
	}
	dst := sourcesAndDest[len(sourcesAndDest)-1]
	src := sourcesAndDest[0 : len(sourcesAndDest)-2]
	return src, dst
}

func extractKeyValuePairsAsMap(pairs instructions.KeyValuePairs) (res map[string]string) {
	res = make(map[string]string)
	for _, item := range pairs {
		res[item.Key] = item.Value
	}
	return
}
