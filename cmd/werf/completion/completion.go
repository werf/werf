package completion

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
)

var cmdData struct {
	Shell string
}

const zshCompdef = "compdef _werf werf\n"

func NewCmd(ctx context.Context, rootCmd *cobra.Command) *cobra.Command {
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "completion",
		DisableFlagsInUseLine: true,
		Short:                 "Generate bash completion scripts",
		Example: fmt.Sprintf(`  # Load bash completion
  $ source <(%[1]s completion)
  # or for older bash versions (e.g. bash 3.2 on macOS):
  $ source /dev/stdin <<< "$(%[1]s completion)"

  # Load zsh completion
  $ autoload -Uz compinit && compinit -C
  $ source <(%[1]s completion --shell=zsh)

  # Load fish completion
  $ source <(%[1]s completion --shell=fish)

  # Load powershell completion
  $ %[1]s completion --shell=powershell | Out-String | Invoke-Expression`, rootCmd.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cmdData.Shell {
			case "bash":
				return rootCmd.GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
					return err
				}

				_, _ = os.Stdout.WriteString(zshCompdef)

				return nil
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				common.PrintHelp(cmd)
				return fmt.Errorf("provided shell %q not supported", cmdData.Shell)
			}
		},
	})

	var defaultShell string
	if os.Getenv("WERF_SHELL") != "" {
		defaultShell = os.Getenv("WERF_SHELL")
	} else {
		defaultShell = "bash"
	}

	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", defaultShell, "Set to bash, zsh, fish or powershell (default $WERF_SHELL or bash)")

	return cmd
}
