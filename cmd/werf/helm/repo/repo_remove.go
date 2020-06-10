package repo

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

type repoRemoveCmd struct {
	out  io.Writer
	name string
	home helmpath.Home
}

func newRepoRemoveCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	remove := &repoRemoveCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "remove [NAME]",
		Short:                 "Remove a chart repository",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			if len(args) == 0 {
				return fmt.Errorf("need at least one argument, name of chart repository")
			}

			remove.home = helmCommon.HelmSettings.Home
			for i := 0; i < len(args); i++ {
				remove.name = args[i]
				if err := remove.run(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (r *repoRemoveCmd) run() error {
	return removeRepoLine(r.out, r.name, r.home)
}

func removeRepoLine(out io.Writer, name string, home helmpath.Home) error {
	repoFile := home.RepositoryFile()
	r, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		if helmCommon.IsCouldNotLoadRepositoriesFileError(err) {
			return fmt.Errorf(helmCommon.CouldNotLoadRepositoriesFileErrorFormat, home.RepositoryFile())
		}

		return err
	}

	if !r.Remove(name) {
		return fmt.Errorf("no repo named %q found", name)
	}
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	if err := removeRepoCache(name, home); err != nil {
		return err
	}

	fmt.Fprintf(out, "%q has been removed from your repositories\n", name)

	return nil
}

func removeRepoCache(name string, home helmpath.Home) error {
	if _, err := os.Stat(home.CacheIndex(name)); err == nil {
		err = os.Remove(home.CacheIndex(name))
		if err != nil {
			return err
		}
	}
	return nil
}
