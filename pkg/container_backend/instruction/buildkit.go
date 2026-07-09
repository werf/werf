package instruction

import (
	"context"
	"strings"

	"github.com/werf/werf/v2/pkg/buildkit"
	"github.com/werf/werf/v2/pkg/dockerfile"
)

func (i *Env) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	for _, item := range i.Env {
		stage.AddEnv(item.Key, item.Value)
	}
	return nil
}

func (i *Label) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	if stage.Image.Config.Labels == nil {
		stage.Image.Config.Labels = map[string]string{}
	}
	for _, item := range i.Labels {
		stage.Image.Config.Labels[item.Key] = item.Value
	}
	return nil
}

func (i *Expose) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	if stage.Image.Config.ExposedPorts == nil {
		stage.Image.Config.ExposedPorts = map[string]struct{}{}
	}
	for _, port := range i.Ports {
		if !strings.Contains(port, "/") {
			port += "/tcp"
		}
		stage.Image.Config.ExposedPorts[port] = struct{}{}
	}
	return nil
}

func (i *Volume) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	if stage.Image.Config.Volumes == nil {
		stage.Image.Config.Volumes = map[string]struct{}{}
	}
	for _, volume := range i.Volumes {
		stage.Image.Config.Volumes[volume] = struct{}{}
	}
	return nil
}

func (i *Workdir) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	return stage.SetWorkdir(ctx, i.Path)
}

func (i *User) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.SetUser(i.User)
	return nil
}

func (i *Cmd) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.Cmd = stage.WithShell(i.CmdLine, i.PrependShell)
	return nil
}

func (i *Entrypoint) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.Entrypoint = stage.WithShell(i.CmdLine, i.PrependShell)
	if i.EntrypointResetCMD {
		stage.Image.Config.Cmd = nil
	}
	return nil
}

func (i *Healthcheck) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.Healthcheck = i.Health
	return nil
}

func (i *OnBuild) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.OnBuild = append(stage.Image.Config.OnBuild, i.Expression)
	return nil
}

func (i *Shell) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.Shell = i.Shell
	return nil
}

func (i *StopSignal) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Config.StopSignal = i.Signal
	return nil
}

func (i *Maintainer) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	stage.Image.Author = i.Maintainer
	return nil
}

func (i *Run) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	cmdLine := i.CmdLine
	prependShell := i.PrependShell
	if len(i.Files) > 0 {
		full, prepend := dockerfile.MapToCorrectHeredocCmd(i.ShellDependantCmdLine)
		cmdLine = []string{full}
		prependShell = prepend
	}

	stage.Secrets = append(stage.Secrets, i.Secrets...)
	if i.SSH != "" {
		stage.SSH = i.SSH
	}

	return stage.RunCommand(stage.WithShell(cmdLine, prependShell), buildkit.RunCommandOptions{
		Envs:     i.Envs,
		Insecure: i.GetSecurity() == "insecure",
		Network:  i.GetNetwork(),
		Mounts:   i.GetMounts(),
	})
}

func (i *Copy) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	return stage.Copy(ctx, i.SourcesAndDest, buildkit.CopyOptions{
		From:  i.From,
		Chown: i.Chown,
		Chmod: i.Chmod,
	})
}

func (i *Add) ApplyBuildkit(ctx context.Context, stage *buildkit.DockerfileStageState) error {
	return stage.Copy(ctx, i.SourcesAndDest, buildkit.CopyOptions{
		Chown: i.Chown,
		Chmod: i.Chmod,
		IsAdd: true,
	})
}
