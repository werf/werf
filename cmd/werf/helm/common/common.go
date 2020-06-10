package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/tag_strategy"
)

func GetEnvironmentOrStub(environmentOption string) string {
	if environmentOption == "" {
		return "ENV"
	}
	return environmentOption
}

func GetTagOrStub(commonCmdData *common.CmdData) (string, tag_strategy.TagStrategy, error) {
	tag, tagStrategy, err := common.GetDeployTag(commonCmdData, common.TagOptionsGetterOptions{Optional: true})
	if err != nil {
		return "", "", err
	}

	if tag == "" {
		tag, tagStrategy = "TAG", tag_strategy.Custom
	}

	return tag, tagStrategy, nil
}

var (
	HelmSettings *helm_env.EnvSettings
)

type HelmCmdData struct {
	helmSettingsHome *string
}

func SetupHelmHome(cmdData *HelmCmdData, cmd *cobra.Command) {
	cmdData.helmSettingsHome = new(string)

	var helmHomeDefaultValue string
	for _, envName := range []string{"WERF_HELM_HOME", "HELM_HOME"} {
		helmHomeDefaultValue = os.Getenv(envName)
		if helmHomeDefaultValue != "" {
			break
		}
	}

	if helmHomeDefaultValue == "" {
		helmHomeDefaultValue = helm_env.DefaultHelmHome
	}

	cmd.Flags().StringVarP(cmdData.helmSettingsHome, "helm-home", "", helmHomeDefaultValue, "location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm")
}

func InitHelmSettings(helmCmdData *HelmCmdData) {
	HelmSettings = new(helm_env.EnvSettings)
	HelmSettings.Home = helmpath.Home(*helmCmdData.helmSettingsHome)
}

func DefaultKeyring() string {
	return "$HOME/.gnupg/pubring.gpg"
}

func CheckArgsLength(argsReceived int, requiredArgs ...string) error {
	expectedNum := len(requiredArgs)
	if argsReceived != expectedNum {
		arg := "arguments"
		if expectedNum == 1 {
			arg = "argument"
		}
		return fmt.Errorf("this command needs %v %s: %s", expectedNum, arg, strings.Join(requiredArgs, ", "))
	}
	return nil
}

var CouldNotLoadRepositoriesFileErrorFormat = "could not load repositories file (%s): you might need to run `werf helm repo init`"

func IsCouldNotLoadRepositoriesFileError(err error) bool {
	return strings.HasPrefix(err.Error(), "Couldn't load repositories file")
}

type DownloadChartOptions struct {
	Untar    bool
	UntarDir string
	ChartRef string
	DestDir  string
	Version  string
	RepoURL  string
	Username string
	Password string

	Verify      bool
	VerifyLater bool
	Keyring     string

	CertFile string
	KeyFile  string
	CaFile   string

	Devel bool

	Out io.Writer
}

func DownloadChart(opts *DownloadChartOptions) error {
	c := downloader.ChartDownloader{
		HelmHome: HelmSettings.Home,
		Out:      opts.Out,
		Keyring:  opts.Keyring,
		Verify:   downloader.VerifyNever,
		Getters:  getter.All(*HelmSettings),
		Username: opts.Username,
		Password: opts.Password,
	}

	if opts.Verify {
		c.Verify = downloader.VerifyAlways
	} else if opts.VerifyLater {
		c.Verify = downloader.VerifyLater
	}
	// If untar is set, we fetch to a tempdir, then untar and copy after
	// verification.
	dest := opts.DestDir
	if opts.Untar {
		var err error
		dest, err = ioutil.TempDir("", "helm-")
		if err != nil {
			return fmt.Errorf("Failed to untar: %s", err)
		}
		defer os.RemoveAll(dest)
	}
	if opts.RepoURL != "" {
		chartURL, err := repo.FindChartInAuthRepoURL(opts.RepoURL, opts.Username, opts.Password, opts.ChartRef, opts.Version, opts.CertFile, opts.KeyFile, opts.CaFile, getter.All(*HelmSettings))
		if err != nil {
			return err
		}
		opts.ChartRef = chartURL
	}
	saved, v, err := c.DownloadTo(opts.ChartRef, opts.Version, dest)
	if err != nil {
		if IsCouldNotLoadRepositoriesFileError(err) {
			return fmt.Errorf(CouldNotLoadRepositoriesFileErrorFormat, c.HelmHome.RepositoryFile())
		}

		if isNoChartVersionFoundError(err) {
			return processNoChartVersionFoundError(err)
		}

		return err
	}
	if opts.Verify {
		fmt.Fprintf(opts.Out, "Verification: %v\n", v)
	}
	// After verification, untar the chart into the requested directory.
	if opts.Untar {
		ud := opts.UntarDir
		if !filepath.IsAbs(ud) {
			ud = filepath.Join(opts.DestDir, ud)
		}
		if fi, err := os.Stat(ud); err != nil {
			if err := os.MkdirAll(ud, 0755); err != nil {
				return fmt.Errorf("Failed to untar (mkdir): %s", err)
			}

		} else if !fi.IsDir() {
			return fmt.Errorf("Failed to untar: %s is not a directory", ud)
		}

		return chartutil.ExpandFile(ud, saved)
	}
	return nil
}

func isNoChartVersionFoundError(err error) bool {
	return strings.Contains(err.Error(), "matching version") && strings.Contains(err.Error(), "not found in")
}

func processNoChartVersionFoundError(err error) error {
	return fmt.Errorf(strings.Replace(err.Error(), "helm repo update", "werf helm repo update", -1))
}
