package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate bash completion scripts",
		Long: fmt.Sprintf(`To load completion run

. <(%[1]s completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(%[1]s completion)
`, rootCmd.Name()),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	return cmd
}
