package repo

import (
	"github.com/spf13/cobra"
)

var repoHelm = `
This command consists of multiple subcommands to interact with chart repositories.
It can be used:
* to init, add, update, remove and list chart repositories 
* to search and fetch charts
`

func NewRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "repo init|add|update|remove|list|fetch|search",
		Short:                 "Work with chart repositories",
		Long:                  repoHelm,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(newRepoInitCmd())
	cmd.AddCommand(newRepoAddCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoUpdateCmd())
	cmd.AddCommand(newRepoRemoveCmd())
	cmd.AddCommand(newRepoSearchCmd())
	cmd.AddCommand(newRepoFetchCmd())

	return cmd
}
