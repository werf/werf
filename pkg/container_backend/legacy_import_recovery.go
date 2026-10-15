package container_backend

import (
	"context"
	"errors"
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/cli/cli"
	"github.com/moby/moby/client"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/docker"
)

func (c *LegacyStageImageContainer) runWithImportRecovery(ctx context.Context) error {
	run := func(ctx context.Context) error {
		if debugDockerRunCommand() {
			args, err := c.prepareRunArgs(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("Docker run command (stage script piped to stdin):\n%s\n", c.prepareDebugRunCommand(ctx, args))
		}
		return c.run(ctx)
	}
	if c.prepareImports == nil {
		return run(ctx)
	}
	info, err := docker.ImageInspect(ctx, c.imageRef(c.image.fromImage))
	if err != nil {
		return fmt.Errorf("inspect import destination base: %w", err)
	}
	c.baseImageID = info.ID
	prepare := func(ctx context.Context) error {
		commands, err := c.prepareImports(ctx)
		if err != nil {
			return fmt.Errorf("prepare import commands: %w", err)
		}
		c.importCommands = commands
		return nil
	}
	return c.runImportAttempts(ctx, prepare, run, func(ctx context.Context) error {
		err := docker.ContainerRemove(ctx, c.name, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
		if err != nil && !cerrdefs.IsNotFound(err) {
			return fmt.Errorf("remove failed import container: %w", err)
		}
		return nil
	})
}

func (c *LegacyStageImageContainer) runImportAttempts(ctx context.Context, prepare, run, discard func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := prepare(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.importDestinationDiscarded = false
	original := run(ctx)
	if original == nil || ctx.Err() != nil || len(c.runCommands) != 0 || len(c.serviceRunCommands) != 0 || !isImportTransportError(original) {
		return original
	}
	if err := discard(ctx); err != nil {
		return errors.Join(original, fmt.Errorf("discard import destination: %w", err))
	}
	c.importDestinationDiscarded = true
	if ctx.Err() != nil {
		return original
	}
	if err := prepare(ctx); err != nil {
		return errors.Join(original, fmt.Errorf("refresh imports: %w", err))
	}
	if ctx.Err() != nil {
		return original
	}
	logboek.Context(ctx).Warn().LogLn("Retrying import stage in a clean container after a transport failure")
	if err := run(ctx); err != nil {
		return errors.Join(original, err)
	}
	return nil
}

func isImportTransportError(err error) bool {
	var status cli.StatusError
	if !errors.As(err, &status) {
		return false
	}
	switch status.StatusCode {
	case 10, 12, 30, 35:
		return true
	default:
		return false
	}
}
