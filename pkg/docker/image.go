package docker

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"

	"github.com/docker/cli/cli/command/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/werf/logboek"
	"golang.org/x/net/context"
)

func CreateImage(ctx context.Context, ref string, labels map[string]string) error {
	var opts types.ImageImportOptions

	if len(labels) > 0 {
		changeOption := "LABEL"

		for k, v := range labels {
			changeOption += fmt.Sprintf(" %s=%s", k, v)
		}

		opts.Changes = append(opts.Changes, changeOption)
	}

	apiClient, err := apiCli()
	if err != nil {
		return err
	}

	_, err = apiClient.ImageImport(ctx, types.ImageImportSource{SourceName: "-"}, ref, opts)
	return err
}

func Images(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	apiClient, err := apiCli()
	if err != nil {
		return nil, err
	}

	images, err := apiClient.ImageList(ctx, options)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func ImageExist(ctx context.Context, ref string) (bool, error) {
	if _, err := ImageInspect(ctx, ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ImageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	apiClient, err := apiCli()
	if err != nil {
		return nil, err
	}

	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

func doCliPull(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewPullCommand(c), args...).Execute()
}

func CliPull(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliPull(c, args...)
	})
}

func CliPull_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliPull(cli, args...)
}

func CliPull_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPull(c, args...)
	})
}

const cliPullMaxAttempts = 5

func doCliPullWithRetries(ctx context.Context, c *command.DockerCli, args ...string) error {
	var attempt int

tryPull:
	if err := doCliPull(c, args...); err != nil {
		if attempt < cliPullMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
				"504 Gateway Time-out",
				"504 Gateway Timeout",
				"Internal Server Error",
			}

			for _, specificError := range specificErrors {
				if strings.Index(err.Error(), specificError) != -1 {
					attempt += 1
					seconds := rand.Intn(30-15) + 15 // from 15 to 30 seconds

					logboek.Context(ctx).Warn().LogF("Retrying docker pull in %d seconds (%d/%d) ...\n", seconds, attempt, cliPullMaxAttempts)
					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPull
				}
			}
		}

		return err
	}

	return nil
}

func CliPullWithRetries(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliPullWithRetries(ctx, c, args...)
	})
}

func CliPullWithRetries_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliPullWithRetries(ctx, cli, args...)
}

func CliPullWithRetries_RecordedOutput(ctx context.Context, args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPullWithRetries(ctx, c, args...)
	})
}

func doCliPush(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewPushCommand(c), args...).Execute()
}

func CliPush(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliPush(c, args...)
	})
}

func CliPush_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliPush(cli, args...)
}

func CliPush_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPush(c, args...)
	})
}

const cliPushMaxAttempts = 10

func doCliPushWithRetries(c *command.DockerCli, args ...string) error {
	var attempt int

tryPush:
	if err := doCliPush(c, args...); err != nil {
		if attempt < cliPushMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
				"Only schema version 2 is supported",
				"504 Gateway Time-out",
				"504 Gateway Timeout",
				"Internal Server Error",
			}

			for _, specificError := range specificErrors {
				if strings.Index(err.Error(), specificError) != -1 {
					attempt += 1
					seconds := rand.Intn(30-15) + 15 // from 15 to 30 seconds

					msg := fmt.Sprintf("Retrying docker push in %d seconds (%d/%d) ...\n", seconds, attempt, cliPushMaxAttempts)
					_, _ = c.Err().Write([]byte(msg))

					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPush
				}
			}
		}

		return err
	}

	return nil
}

func CliPushWithRetries(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliPushWithRetries(c, args...)
	})
}

func CliPushWithRetries_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliPushWithRetries(cli, args...)
}

func CliPushWithRetries_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPushWithRetries(c, args...)
	})
}

func doCliTag(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewTagCommand(c), args...).Execute()
}

func CliTag(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliTag(c, args...)
	})
}

func CliTag_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliTag(cli, args...)
}

func CliTag_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliTag(c, args...)
	})
}

func doCliRmi(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewRemoveCommand(c), args...).Execute()
}

func CliRmi(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliRmi(c, args...)
	})
}

func CliRmi_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliRmi(cli, args...)
}

func CliRmiOutput_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRmi(c, args...)
	})
}

func doCliBuild(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewBuildCommand(c), args...).Execute()
}

func CliBuild(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c *command.DockerCli) error {
		return doCliBuild(c, args...)
	})
}

func CliBuild_LiveOutput(ctx context.Context, args ...string) error {
	cli, err := cli(ctx)
	if err != nil {
		return err
	}

	return doCliBuild(cli, args...)
}

func CliBuild_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliBuild(c, args...)
	})
}
