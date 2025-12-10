package instruction

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/buildah"
	"github.com/werf/werf/v2/pkg/container_backend"
)

type Run struct {
	instructions.RunCommand
	Envs    []string
	Secrets []string
	SSH     string
}

func NewRun(i instructions.RunCommand, envs, secrets []string, ssh string) *Run {
	return &Run{RunCommand: i, Envs: envs, Secrets: secrets, SSH: ssh}
}

func (i *Run) UsesBuildContext() bool {
	for _, mount := range i.GetMounts() {
		if mount.Type == instructions.MountTypeBind && mount.From == "" {
			return true
		}
	}

	return false
}

func (i *Run) GetMounts() []*instructions.Mount {
	return instructions.GetMounts(&i.RunCommand)
}

func (i *Run) GetSecurity() string {
	return instructions.GetSecurity(&i.RunCommand)
}

func (i *Run) GetNetwork() string {
	return instructions.GetNetwork(&i.RunCommand)
}

func (i *Run) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	var contextDir string
	if i.UsesBuildContext() {
		var err error
		contextDir, err = buildContextArchive.ExtractOrGetExtractedDir(ctx)
		if err != nil {
			return fmt.Errorf("unable to extract build context: %w", err)
		}
	}

	var addCapabilities []string
	if i.GetSecurity() == "insecure" {
		addCapabilities = []string{"all"}
	}

	cmdLine := i.CmdLine
	prependShell := i.PrependShell
	if len(i.Files) > 0 {
		f := i.Files[0]

		script := f.Data
		if f.Chomp {
			script = strings.TrimRight(script, "\r\n")
		}

		cmdLine = []string{"sh", "-c", script}
		prependShell = false
	}

	logboek.Context(ctx).Default().LogF("$ %s\n", strings.Join(cmdLine, " "))

	if err := drv.RunCommand(ctx, containerName, cmdLine, buildah.RunCommandOpts{
		CommonOpts:      drvOpts,
		ContextDir:      contextDir,
		PrependShell:    prependShell,
		AddCapabilities: addCapabilities,
		NetworkType:     i.GetNetwork(),
		RunMounts:       i.GetMounts(),
		Envs:            i.Envs,
		Secrets:         i.Secrets,
		SSH:             i.SSH,
	}); err != nil {
		return fmt.Errorf("error running command %v for container %s: %w", cmdLine, containerName, err)
	}

	return nil
}
