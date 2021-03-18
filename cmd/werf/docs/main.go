package docs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/cmd/werf/common/templates"
)

var commonCmdData common.CmdData

func NewCmd(cmdGroups *templates.CommandGroups) *cobra.Command {
	cmd := &cobra.Command{
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

			partialsDir := filepath.Join(workingDir, "docs/documentation/_includes/reference/cli")
			pagesDir := filepath.Join(workingDir, "docs/documentation/pages_en/reference/cli")
			sidebarPath := filepath.Join(workingDir, "docs/documentation/_data/sidebars/_cli.yml")

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
	}

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupDir(&commonCmdData, cmd)

	return cmd
}

func createEmptyFolder(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("unable to remove %s: %s", path, err)
	}

	if err := os.MkdirAll(path, 0777); err != nil {
		return fmt.Errorf("unable to make dir %s: %s", path, err)
	}

	return nil
}
