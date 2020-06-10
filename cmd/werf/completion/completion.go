package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
)

var cmdData struct {
	Shell string
}

func NewCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "completion",
		DisableFlagsInUseLine: true,
		Short:                 "Generate bash completion scripts",
		Example: fmt.Sprintf(`  # Load bash completion
  $ source <(%[1]s completion)

  # Load zsh completion
  $ source <(%[1]s completion --shell=zsh)`, rootCmd.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cmdData.Shell {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			default:
				common.PrintHelp(cmd)
				return fmt.Errorf("provided shell '%s' not supported", cmdData.Shell)
			}
		},
	}

	var defaultShell string
	if os.Getenv("WERF_SHELL") != "" {
		defaultShell = os.Getenv("WERF_SHELL")
	} else {
		defaultShell = "bash"
	}

	cmd.Flags().StringVarP(&cmdData.Shell, "shell", "", defaultShell, "Set to bash or zsh (default $WERF_SHELL or bash)")

	return cmd
}
