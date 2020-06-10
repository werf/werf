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

	"github.com/werf/werf/cmd/werf/common"
	helm_common "github.com/werf/werf/cmd/werf/helm/common"
)

const dependencyUpDesc = `
Update the on-disk dependencies to mirror the requirements.yaml file.

This command verifies that the required charts, as expressed in 'requirements.yaml',
are present in 'charts/' and are at an acceptable version. It will pull down
the latest charts that satisfy the dependencies, and clean up old dependencies.

On successful update, this will generate a lock file that can be used to
rebuild the requirements to an exact version.

Dependencies are not required to be represented in 'requirements.yaml'. For that
reason, an update command will not remove charts unless they are (a) present
in the requirements.yaml file, but (b) at the wrong version.
`

// dependencyUpdateCmd describes a 'helm dependency update'
type dependencyUpdateCmd struct {
	out         io.Writer
	chartpath   string
	helmhome    helmpath.Home
	verify      bool
	keyring     string
	skipRefresh bool
}

// newDependencyUpdateCmd creates a new dependency update command.
func newDependencyUpdateCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helm_common.HelmCmdData
	duc := &dependencyUpdateCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "update",
		Short:                 "Update charts/ based on the contents of requirements.yaml",
		Long:                  dependencyUpDesc,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helm_common.InitHelmSettings(&helmCommonCmdData)

			projectDir, err := common.GetProjectDir(&commonCmdData)
			if err != nil {
				return fmt.Errorf("getting project dir failed: %s", err)
			}

			helmChartDir, err := common.GetHelmChartDir(projectDir, &commonCmdData)
			if err != nil {
				return fmt.Errorf("getting helm chart dir failed: %s", err)
			}

			duc.chartpath = helmChartDir
			duc.helmhome = helm_common.HelmSettings.Home

			return duc.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&duc.verify, "verify", false, "verify the packages against signatures")
	f.StringVar(&duc.keyring, "keyring", helm_common.DefaultKeyring(), "keyring containing public keys")
	f.BoolVar(&duc.skipRefresh, "skip-refresh", false, "do not refresh the local repository cache")

	common.SetupDir(&commonCmdData, cmd)
	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupHelmChartDir(&commonCmdData, cmd)

	helm_common.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

// run runs the full dependency update process.
func (d *dependencyUpdateCmd) run() error {
	man := &downloader.Manager{
		Out:        d.out,
		ChartPath:  d.chartpath,
		HelmHome:   d.helmhome,
		Keyring:    d.keyring,
		SkipUpdate: d.skipRefresh,
		Getters:    getter.All(*helm_common.HelmSettings),
	}
	if d.verify {
		man.Verify = downloader.VerifyAlways
	}
	if helm_common.HelmSettings.Debug {
		man.Debug = true
	}

	return chartutil.WithSkipChartYamlFileValidation(true, func() error {
		if err := man.Update(); err != nil {
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
