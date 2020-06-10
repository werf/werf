package repo

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

type repoAddCmd struct {
	name     string
	url      string
	username string
	password string
	home     helmpath.Home
	noupdate bool

	certFile string
	keyFile  string
	caFile   string

	out io.Writer
}

func newRepoAddCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	add := &repoAddCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "add [NAME] [URL]",
		Short:                 "Add a chart repository",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			if err := helmCommon.CheckArgsLength(len(args), "name for the chart repository", "the url of the chart repository"); err != nil {
				return err
			}

			add.name = args[0]
			add.url = args[1]
			add.home = helmCommon.HelmSettings.Home

			return add.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&add.username, "username", "", "chart repository username")
	f.StringVar(&add.password, "password", "", "chart repository password")
	f.BoolVar(&add.noupdate, "no-update", false, "raise error if repo is already registered")
	f.StringVar(&add.certFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&add.keyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&add.caFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (a *repoAddCmd) run() error {
	if a.username != "" && a.password == "" {
		fmt.Fprint(a.out, "Password:")
		password, err := readPassword()
		fmt.Fprintln(a.out)
		if err != nil {
			return err
		}
		a.password = password
	}

	if err := addRepository(a.name, a.url, a.username, a.password, a.home, a.certFile, a.keyFile, a.caFile, a.noupdate); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "%q has been added to your repositories\n", a.name)
	return nil
}

func readPassword() (string, error) {
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func addRepository(name, url, username, password string, home helmpath.Home, certFile, keyFile, caFile string, noUpdate bool) error {
	f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		if helmCommon.IsCouldNotLoadRepositoriesFileError(err) {
			return fmt.Errorf(helmCommon.CouldNotLoadRepositoriesFileErrorFormat, home.RepositoryFile())
		}

		return err
	}

	if noUpdate && f.Has(name) {
		return fmt.Errorf("repository name (%s) already exists, please specify a different name", name)
	}

	cif := home.CacheIndex(name)
	c := repo.Entry{
		Name:     name,
		Cache:    cif,
		URL:      url,
		Username: username,
		Password: password,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}

	r, err := repo.NewChartRepository(&c, getter.All(*helmCommon.HelmSettings))
	if err != nil {
		return err
	}

	if err := r.DownloadIndexFile(home.Cache()); err != nil {
		return fmt.Errorf("looks like %q is not a valid chart repository or cannot be reached: %s", url, err.Error())
	}

	f.Update(&c)

	return f.WriteFile(home.RepositoryFile(), 0644)
}
