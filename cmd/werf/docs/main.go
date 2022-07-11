package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/common/templates"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context, cmdGroups *templates.CommandGroups) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, &cobra.Command{
		Use:                   "docs",
		DisableFlagsInUseLine: true,
		Short:                 "Generate documentation as markdown",
		Hidden:                true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			workingDir := common.GetWorkingDir(&commonCmdData)

			partialsDir := filepath.Join(workingDir, "docs/_includes/reference/cli")
			pagesDir := filepath.Join(workingDir, "docs/pages_en/reference/cli")
			sidebarPath := filepath.Join(workingDir, "docs/_data/sidebars/_cli.yml")

			for _, path := range []string{partialsDir, pagesDir} {
				if err := createEmptyFolder(path); err != nil {
					return err
				}
			}

			if err := GenCliPartials(cmd.Root(), partialsDir); err != nil {
				return err
			}

			if err := GenCliOverview(*cmdGroups, pagesDir); err != nil {
				return err
			}

			if err := GenCliPages(*cmdGroups, pagesDir); err != nil {
				return err
			}

			if err := GenCliSidebar(*cmdGroups, sidebarPath); err != nil {
				return err
			}

			return nil
		},
	})

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupDir(&commonCmdData, cmd)

	return cmd
}

func createEmptyFolder(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("unable to remove %s: %w", path, err)
	}

	if err := os.MkdirAll(path, 0o777); err != nil {
		return fmt.Errorf("unable to make dir %s: %w", path, err)
	}

	return nil
}
