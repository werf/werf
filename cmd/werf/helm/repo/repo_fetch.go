package repo

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

const fetchDesc = `
Retrieve a package from a package repository, and download it locally.

This is useful for fetching packages to inspect, modify, or repackage. It can
also be used to perform cryptographic verification of a chart without installing
the chart.

There are options for unpacking the chart after download. This will create a
directory for the chart and uncompress into that directory.

If the --verify flag is specified, the requested chart MUST have a provenance
file, and MUST pass the verification process. Failure in any part of this will
result in an error, and the chart will not be saved locally.
`

func newRepoFetchCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	downloadChartOptions := &helmCommon.DownloadChartOptions{Out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "fetch [chart URL | repo/chartname] [...]",
		Short:                 "Download a chart from a repository and (optionally) unpack it in local directory",
		Long:                  fetchDesc,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			if downloadChartOptions.Keyring != "" {
				downloadChartOptions.Keyring = os.ExpandEnv(downloadChartOptions.Keyring)
			}

			if len(args) == 0 {
				return fmt.Errorf("need at least one argument, url or repo/name of the chart")
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			if downloadChartOptions.Version == "" && downloadChartOptions.Devel {
				downloadChartOptions.Version = ">0.0.0-0"
			}

			for i := 0; i < len(args); i++ {
				downloadChartOptions.ChartRef = args[i]
				if err := helmCommon.DownloadChart(downloadChartOptions); err != nil {
					return err
				}
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVar(&downloadChartOptions.Untar, "untar", false, "if set to true, will untar the chart after downloading it")
	f.StringVar(&downloadChartOptions.UntarDir, "untardir", ".", "if untar is specified, this flag specifies the name of the directory into which the chart is expanded")
	f.BoolVar(&downloadChartOptions.Verify, "verify", false, "verify the package against its signature")
	f.BoolVar(&downloadChartOptions.VerifyLater, "prov", false, "fetch the provenance file, but don't perform verification")
	f.StringVar(&downloadChartOptions.Version, "version", "", "specific version of a chart. Without this, the latest version is fetched")
	f.StringVar(&downloadChartOptions.Keyring, "keyring", helmCommon.DefaultKeyring(), "keyring containing public keys")
	f.StringVarP(&downloadChartOptions.DestDir, "destination", "d", ".", "location to write the chart. If this and tardir are specified, tardir is appended to this")
	f.StringVar(&downloadChartOptions.RepoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&downloadChartOptions.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&downloadChartOptions.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&downloadChartOptions.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&downloadChartOptions.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored.")
	f.StringVar(&downloadChartOptions.Username, "username", "", "chart repository username")
	f.StringVar(&downloadChartOptions.Password, "password", "", "chart repository password")

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}
