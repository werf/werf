package repo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

const updateDesc = `
Update gets the latest information about charts from the respective chart repositories.
Information is cached locally, where it is used by commands like 'werf helm repo search'.
`

var errNoRepositories = errors.New("no repositories found. You must add one before updating")

type repoUpdateCmd struct {
	update func([]*repo.ChartRepository, io.Writer, helmpath.Home, bool) error
	home   helmpath.Home
	out    io.Writer
	strict bool
}

func newRepoUpdateCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	u := &repoUpdateCmd{
		out:    os.Stdout,
		update: updateCharts,
	}

	cmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Update information of available charts locally from chart repositories",
		Long:                  updateDesc,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			u.home = helmCommon.HelmSettings.Home
			return u.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&u.strict, "strict", false, "fail on update warnings")

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (u *repoUpdateCmd) run() error {
	f, err := repo.LoadRepositoriesFile(u.home.RepositoryFile())
	if err != nil {
		if helmCommon.IsCouldNotLoadRepositoriesFileError(err) {
			return fmt.Errorf(helmCommon.CouldNotLoadRepositoriesFileErrorFormat, u.home.RepositoryFile())
		}

		return err
	}

	if len(f.Repositories) == 0 {
		return errNoRepositories
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(*helmCommon.HelmSettings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}
	return u.update(repos, u.out, u.home, u.strict)
}

func updateCharts(repos []*repo.ChartRepository, out io.Writer, home helmpath.Home, strict bool) error {
	fmt.Fprintln(out, "Hang tight while we grab the latest from your chart repositories...")
	var (
		errorCounter int
		wg           sync.WaitGroup
		mu           sync.Mutex
	)
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if re.Config.Name == installer.LocalRepository {
				mu.Lock()
				fmt.Fprintf(out, "...Skip %s chart repository\n", re.Config.Name)
				mu.Unlock()
				return
			}
			err := re.DownloadIndexFile(home.Cache())
			if err != nil {
				mu.Lock()
				errorCounter++
				fmt.Fprintf(out, "...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
				mu.Unlock()
			} else {
				mu.Lock()
				fmt.Fprintf(out, "...Successfully got an update from the %q chart repository\n", re.Config.Name)
				mu.Unlock()
			}
		}(re)
	}
	wg.Wait()

	if errorCounter != 0 && strict {
		return errors.New("Update Failed. Check log for details")
	}

	fmt.Fprintln(out, "Update Complete.")
	return nil
}
