package docker

import (
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

func CreateImage(ref string) error {
	ctx := context.Background()
	_, err := apiClient.ImageImport(ctx, types.ImageImportSource{SourceName: "-"}, ref, types.ImageImportOptions{})
	return err
}

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

func doCliPull(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewPullCommand(c), args...).Execute()
}

func CliPull(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliPull(c, args...)
	})
}

func CliPull_LiveOutput(args ...string) error {
	return doCliPull(liveOutputCli, args...)
}

func CliPull_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPull(c, args...)
	})
}

const cliPullMaxAttempts = 5

func doCliPullWithRetries(c *command.DockerCli, args ...string) error {
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

					logboek.LogWarnF("Retrying docker pull in %d seconds (%d/%d) ...\n", seconds, attempt, cliPullMaxAttempts)
					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPull
				}
			}
		}

		return err
	}

	return nil
}

func CliPullWithRetries(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliPullWithRetries(c, args...)
	})
}

func CliPullWithRetries_LiveOutput(args ...string) error {
	return doCliPullWithRetries(liveOutputCli, args...)
}

func CliPullWithRetries_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPullWithRetries(c, args...)
	})
}

func doCliPush(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewPushCommand(c), args...).Execute()
}

func CliPush(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliPush(c, args...)
	})
}

func CliPush_LiveOutput(args ...string) error {
	return doCliPush(liveOutputCli, args...)
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

					logboek.Warn.LogFDetails("Retrying docker push in %d seconds (%d/%d) ...\n", seconds, attempt, cliPushMaxAttempts)

					time.Sleep(time.Duration(seconds) * time.Second)
					goto tryPush
				}
			}
		}

		return err
	}

	return nil
}

func CliPushWithRetries(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliPushWithRetries(c, args...)
	})
}

func CliPushWithRetries_LiveOutput(args ...string) error {
	return doCliPushWithRetries(liveOutputCli, args...)
}

func CliPushWithRetries_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliPushWithRetries(c, args...)
	})
}

func doCliTag(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewTagCommand(c), args...).Execute()
}

func CliTag(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliTag(c, args...)
	})
}

func CliTag_LiveOutput(args ...string) error {
	return doCliTag(liveOutputCli, args...)
}

func CliTag_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliTag(c, args...)
	})
}

func doCliRmi(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewRemoveCommand(c), args...).Execute()
}

func CliRmi(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliRmi(c, args...)
	})
}

func CliRmi_LiveOutput(args ...string) error {
	return doCliRmi(liveOutputCli, args...)
}

func CliRmiOutput_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliRmi(c, args...)
	})
}

func doCliBuild(c *command.DockerCli, args ...string) error {
	return prepareCliCmd(image.NewBuildCommand(c), args...).Execute()
}

func CliBuild(args ...string) error {
	return callCliWithAutoOutput(func(c *command.DockerCli) error {
		return doCliBuild(c, args...)
	})
}

func CliBuild_LiveOutput(args ...string) error {
	return doCliBuild(liveOutputCli, args...)
}

func CliBuild_RecordedOutput(args ...string) (string, error) {
	return callCliWithRecordedOutput(func(c *command.DockerCli) error {
		return doCliBuild(c, args...)
	})
}
