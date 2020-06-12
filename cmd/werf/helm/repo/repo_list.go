package repo

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

type repoListCmd struct {
	out  io.Writer
	home helmpath.Home
}

func newRepoListCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	list := &repoListCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 "List chart repositories",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			list.home = helmCommon.HelmSettings.Home
			return list.run()
		},
	}

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (a *repoListCmd) run() error {
	f, err := repo.LoadRepositoriesFile(a.home.RepositoryFile())
	if err != nil {
		if helmCommon.IsCouldNotLoadRepositoriesFileError(err) {
			return fmt.Errorf(helmCommon.CouldNotLoadRepositoriesFileErrorFormat, a.home.RepositoryFile())
		}

		return err
	}
	if len(f.Repositories) == 0 {
		return errors.New("no repositories to show")
	}
	table := uitable.New()
	table.AddRow("NAME", "URL")
	for _, re := range f.Repositories {
		table.AddRow(re.Name, re.URL)
	}
	fmt.Fprintln(a.out, table)
	return nil
}
