package logout

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/config/types"
	"github.com/goware/urlx"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
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

			var registry string
			switch {
			case len(args) == 0:
				registry = "https://index.docker.io/v1/"
			case len(args) == 1:
				registry = args[0]
			default: // len(args) > 1
				common.PrintHelp(cmd)
				return fmt.Errorf("invalid number of arguments, expected optional registry address: got %d arguments", len(args))
			}

			return Logout(ctx, registry, LogoutOptions{
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
	u, err := urlx.Parse(registry)
	if err != nil {
		return fmt.Errorf("unable to parse %q: %w", registry, err)
	}

	err = docker.EraseCredentials(opts.DockerConfigDir, types.AuthConfig{
		ServerAddress: u.Host,
	})
	if err != nil {
		return fmt.Errorf("unable to logout from %q: %w", registry, err)
	}

	logboek.Context(ctx).Default().LogFHighlight("Successful logout\n")

	return nil
}
