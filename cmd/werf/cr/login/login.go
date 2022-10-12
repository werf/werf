package login

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"oras.land/oras-go/pkg/auth"
	"oras.land/oras-go/pkg/auth/docker"

	"github.com/werf/logboek"
	"github.com/werf/werf/cmd/werf/common"
	secret_common "github.com/werf/werf/cmd/werf/helm/secret/common"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
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
werf cr login -p token registry.example.com

# Login into insecure registry (over http)
werf cr login --insecure-registry registry.example.com`,
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

			return Login(ctx, args[0], LoginOptions{
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
	var dockerConfigDir string
	if opts.DockerConfigDir != "" {
		dockerConfigDir = opts.DockerConfigDir
	} else {
		dockerConfigDir = filepath.Join(os.Getenv("HOME"), ".docker")
	}

	cli, err := docker.NewClient(filepath.Join(dockerConfigDir, "config.json"))
	if err != nil {
		return fmt.Errorf("unable to create oras auth client: %w", err)
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

	if err := cli.LoginWithOpts(func(settings *auth.LoginSettings) {
		settings.Context = ctx
		settings.Hostname = registry
		settings.Username = opts.Username
		settings.Secret = password
		settings.Insecure = opts.InsecureRegistry
		settings.UserAgent = fmt.Sprintf("werf %s", werf.Version)
	}); err != nil {
		return fmt.Errorf("unable to login into %q: %w", registry, err)
	}

	logboek.Context(ctx).Default().LogFHighlight("Successful login\n")

	return nil
}
