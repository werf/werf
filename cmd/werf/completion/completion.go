package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "completion",
		DisableFlagsInUseLine: true,
		Short: "Generate bash completion scripts",
		Example: fmt.Sprintf(`  # Load completion run
  $ source <(%[1]s completion)`, rootCmd.Name()),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	return cmd
}
