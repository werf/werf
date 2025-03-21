package login

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/config/types"
	"github.com/goware/urlx"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	secret_common "github.com/werf/nelm/pkg/secret"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/docker_registry/auth"
	"github.com/werf/werf/v2/pkg/werf"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

var commonCmdData common.CmdData

var cmdData struct {
	Username      string
	Password      string
	PasswordStdin bool
}

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:   "login registry",
		Short: "Login into remote registry.",
		Long:  common.GetLongCommandDescription(`Login into remote registry.`),
		Example: `# Login with username and password from command line
werf cr login -u username -p password registry.example.com

# Login with token from command line
werf cr login -u username -p token registry.example.com

# Login into insecure registry (over http)
werf cr login -u username -p password --insecure-registry registry.example.com`,
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

			return Login(ctx, registry, LoginOptions{
				Username:         cmdData.Username,
				Password:         cmdData.Password,
				PasswordStdin:    cmdData.PasswordStdin,
				DockerConfigDir:  *commonCmdData.DockerConfig,
				InsecureRegistry: *commonCmdData.InsecureRegistry,
			})
		},
	})

	common.SetupDockerConfig(&commonCmdData, cmd, "")
	common.SetupInsecureRegistry(&commonCmdData, cmd)
	// common.SetupSkipTlsVerifyRegistry(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	cmd.Flags().StringVarP(&cmdData.Username, "username", "u", os.Getenv("WERF_USERNAME"), "Use specified username for login (default $WERF_USERNAME)")
	cmd.Flags().StringVarP(&cmdData.Password, "password", "p", os.Getenv("WERF_PASSWORD"), "Use specified password for login (default $WERF_PASSWORD)")
	cmd.Flags().BoolVarP(&cmdData.PasswordStdin, "password-stdin", "", util.GetBoolEnvironmentDefaultFalse("WERF_PASSWORD_STDIN"), "Read password from stdin for login (default $WERF_PASSWORD_STDIN)")

	return cmd
}

type LoginOptions struct {
	Username         string
	Password         string
	PasswordStdin    bool
	DockerConfigDir  string
	InsecureRegistry bool
}

func Login(ctx context.Context, registry string, opts LoginOptions) error {
	u, err := urlx.Parse(registry)
	if err != nil {
		return fmt.Errorf("unable to parse %q: %w", registry, err)
	}

	if opts.Username == "" {
		return fmt.Errorf("provide --username")
	}

	var password string
	if opts.PasswordStdin {
		if opts.Password != "" {
			return fmt.Errorf("--password and --password-stdin could not be used at the same time")
		}

		var bytePassword []byte
		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			bytePassword, err = secret_common.InputFromInteractiveStdin("Password: ")
			if err != nil {
				return fmt.Errorf("error reading password from interactive stdin: %w", err)
			}
		} else {
			bytePassword, err = secret_common.InputFromStdin()
			if err != nil {
				return fmt.Errorf("error reading password from stdin: %w", err)
			}
		}

		password = string(bytePassword)
	} else if opts.Password != "" {
		password = opts.Password
	} else {
		return fmt.Errorf("provide --password or --password-stdin")
	}

	token, err := auth.Auth(ctx, auth.Options{
		Username:  opts.Username,
		Password:  password,
		UserAgent: werf.UserAgent,
		Hostname:  u.Host,
		Insecure:  opts.InsecureRegistry,
	})
	if err != nil {
		return fmt.Errorf("unable to authenticate into %q: %w", registry, err)
	}

	err = docker.StoreCredentials(opts.DockerConfigDir, types.AuthConfig{
		ServerAddress: u.Host,
		Username:      opts.Username,
		Password:      password,
		IdentityToken: token,
	})
	if err != nil {
		return fmt.Errorf("unable to store credentials: %w", err)
	}

	logboek.Context(ctx).Default().LogFHighlight("Successful login\n")

	return nil
}
