package docker

import (
	"strings"
	"time"

	"github.com/docker/cli/cli/command/image"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/flant/logboek"
	"golang.org/x/net/context"
)

func Images(options types.ImageListOptions) ([]types.ImageSummary, error) {
	ctx := context.Background()
	images, err := apiClient.ImageList(ctx, options)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func ImageExist(ref string) (bool, error) {
	if _, err := ImageInspect(ref); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ImageInspect(ref string) (*types.ImageInspect, error) {
	ctx := context.Background()
	inspect, _, err := apiClient.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &inspect, nil
}

const cliPullMaxAttempts = 5

func CliPullWithRetries(args ...string) error {
	var attempt int

tryPull:
	if err := CliPull(args...); err != nil {
		if attempt < cliPullMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
			}

			for _, specificError := range specificErrors {
				if strings.Index(err.Error(), specificError) != -1 {
					attempt += 1

					logboek.LogWarnF("Retrying in 5 seconds (%d/%d) ...\n", attempt, cliPullMaxAttempts)
					time.Sleep(5 * time.Second)
					goto tryPull
				}
			}
		}

		return err
	}

	return nil
}

func CliPull(args ...string) error {
	cmd := image.NewPullCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliPush(args ...string) error {
	cmd := image.NewPushCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

const cliPushMaxAttempts = 5

func CliPushWithRetries(args ...string) error {
	var attempt int

tryPush:
	if err := CliPush(args...); err != nil {
		if attempt < cliPushMaxAttempts {
			specificErrors := []string{
				"Client.Timeout exceeded while awaiting headers",
				"TLS handshake timeout",
				"i/o timeout",
			}

			for _, specificError := range specificErrors {
				if strings.Index(err.Error(), specificError) != -1 {
					attempt += 1

					logboek.Default.LogFDetails("Retrying in 5 seconds (%d/%d) ...\n", attempt, cliPushMaxAttempts)

					time.Sleep(5 * time.Second)
					goto tryPush
				}
			}
		}

		return err
	}

	return nil
}

func CliTag(args ...string) error {
	cmd := image.NewTagCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliRmi(args ...string) error {
	cmd := image.NewRemoveCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func CliBuild(args ...string) error {
	cmd := image.NewBuildCommand(cli)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	return nil
}
