package dependency

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"

	"github.com/flant/werf/cmd/werf/common"
	helm_common "github.com/flant/werf/cmd/werf/helm/common"
)

const dependencyBuildDesc = `
Build out the charts/ directory from the requirements.lock file.

Build is used to reconstruct a chart's dependencies to the state specified in
the lock file.

If no lock file is found, 'werf helm dependency build' will mirror the behavior of
the 'werf helm dependency update' command. This means it will update the on-disk
dependencies to mirror the requirements.yaml file and generate a lock file.
`

type dependencyBuildCmd struct {
	out       io.Writer
	chartpath string
	verify    bool
	keyring   string
	helmhome  helmpath.Home
}

func newDependencyBuildCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helm_common.HelmCmdData
	dbc := &dependencyBuildCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "build",
		Short:                 "Rebuild the charts/ directory based on the requirements.lock file",
		Long:                  dependencyBuildDesc,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			helm_common.InitHelmSettings(&helmCommonCmdData)

			chartPath, err := getWerfChartPath(commonCmdData)
			if err != nil {
				return err
			}

			dbc.helmhome = helm_common.HelmSettings.Home
			dbc.chartpath = chartPath

			return dbc.run()
		},
	}

	common.SetupDir(&commonCmdData, cmd)

	f := cmd.Flags()
	f.BoolVar(&dbc.verify, "verify", false, "verify the packages against signatures")
	f.StringVar(&dbc.keyring, "keyring", helm_common.DefaultKeyring(), "keyring containing public keys")

	helm_common.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (d *dependencyBuildCmd) run() error {
	man := &downloader.Manager{
		Out:       d.out,
		ChartPath: d.chartpath,
		HelmHome:  d.helmhome,
		Keyring:   d.keyring,
		Getters:   getter.All(*helm_common.HelmSettings),
	}
	if d.verify {
		man.Verify = downloader.VerifyIfPossible
	}

	return chartutil.WithSkipChartYamlFileValidation(func() error {
		if err := man.Build(); err != nil {
			if helm_common.IsCouldNotLoadRepositoriesFileError(err) {
				return fmt.Errorf(helm_common.CouldNotLoadRepositoriesFileErrorFormat, helm_common.HelmSettings.Home.RepositoryFile())
			}

			if isNoRepositoryDefinitionError(err) {
				return processNoRepositoryDefinitionError(err)
			}

			return err
		}

		return nil
	})
}
