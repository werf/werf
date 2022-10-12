package logout

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"oras.land/oras-go/pkg/auth/docker"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "logout registry",
		Short:                 "Logout from a remote registry.",
		Long:                  common.GetLongCommandDescription(`Logout from a remote registry.`),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			defer global_warnings.PrintGlobalWarnings(ctx)

			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if len(args) != 1 {
				common.PrintHelp(cmd)
				return fmt.Errorf("registry address argument required")
			}

			return Logout(ctx, args[0], LogoutOptions{
				DockerConfigDir: *commonCmdData.DockerConfig,
			})
		},
	})

	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

type LogoutOptions struct {
	DockerConfigDir string
}

func Logout(ctx context.Context, registry string, opts LogoutOptions) error {
	var dockerConfigDir string
	if opts.DockerConfigDir != "" {
		dockerConfigDir = opts.DockerConfigDir
	} else {
		dockerConfigDir = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	cli, err := docker.NewClient(filepath.Join(dockerConfigDir, "config.json"))
	if err != nil {
		return fmt.Errorf("unable to create auth client: %w", err)
	}

	if err := cli.Logout(ctx, registry); err != nil {
		return fmt.Errorf("unable to logout from %q: %w", registry, err)
	}

	logboek.Context(ctx).Default().LogFHighlight("Successful logout\n")

	return nil
}
