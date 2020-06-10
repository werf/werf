package repo

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"k8s.io/helm/cmd/helm/search"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/werf/werf/cmd/werf/common"
	helmCommon "github.com/werf/werf/cmd/werf/helm/common"
)

const searchDesc = `
Search reads through all of the repositories configured on the system, and
looks for matches
`

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

type searchCmd struct {
	out      io.Writer
	helmhome helmpath.Home

	versions bool
	regexp   bool
	version  string
	colWidth uint
}

func newRepoSearchCmd() *cobra.Command {
	var commonCmdData common.CmdData
	var helmCommonCmdData helmCommon.HelmCmdData

	sc := &searchCmd{out: os.Stdout}

	cmd := &cobra.Command{
		Use:                   "search [keyword]",
		Short:                 "Search for a keyword in charts",
		Long:                  searchDesc,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			helmCommon.InitHelmSettings(&helmCommonCmdData)

			sc.helmhome = helmCommon.HelmSettings.Home
			return sc.run(args)
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&sc.regexp, "regexp", "r", false, "use regular expressions for searching")
	f.BoolVarP(&sc.versions, "versions", "l", false, "show the long listing, with each version of each chart on its own line")
	f.StringVarP(&sc.version, "version", "v", "", "search using semantic versioning constraints")
	f.UintVar(&sc.colWidth, "col-width", 60, "specifies the max column width of output")

	common.SetupLogOptions(&commonCmdData, cmd)

	helmCommon.SetupHelmHome(&helmCommonCmdData, cmd)

	return cmd
}

func (s *searchCmd) run(args []string) error {
	index, err := s.buildIndex()
	if err != nil {
		return err
	}

	var res []*search.Result
	if len(args) == 0 {
		res = index.All()
	} else {
		q := strings.Join(args, " ")
		res, err = index.Search(q, searchMaxScore, s.regexp)
		if err != nil {
			return err
		}
	}

	search.SortScore(res)
	data, err := s.applyConstraint(res)
	if err != nil {
		return err
	}

	fmt.Fprintln(s.out, strings.TrimSpace(s.formatSearchResults(data, s.colWidth)))

	return nil
}

func (s *searchCmd) applyConstraint(res []*search.Result) ([]*search.Result, error) {
	if len(s.version) == 0 {
		return res, nil
	}

	constraint, err := semver.NewConstraint(s.version)
	if err != nil {
		return res, fmt.Errorf("an invalid version/constraint format: %s", err)
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		if _, found := foundNames[r.Name]; found {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil || constraint.Check(v) {
			data = append(data, r)
			if !s.versions {
				foundNames[r.Name] = true // If user hasn't requested all versions, only show the latest that matches
			}
		}
	}

	return data, nil
}

func (s *searchCmd) formatSearchResults(res []*search.Result, colWidth uint) string {
	if len(res) == 0 {
		return "No results found"
	}
	table := uitable.New()
	table.MaxColWidth = colWidth
	table.AddRow("NAME", "CHART VERSION", "APP VERSION", "DESCRIPTION")
	for _, r := range res {
		table.AddRow(r.Name, r.Chart.Version, r.Chart.AppVersion, r.Chart.Description)
	}
	return table.String()
}

func (s *searchCmd) buildIndex() (*search.Index, error) {
	// Load the repositories.yaml
	rf, err := repo.LoadRepositoriesFile(s.helmhome.RepositoryFile())
	if err != nil {
		if helmCommon.IsCouldNotLoadRepositoriesFileError(err) {
			return nil, fmt.Errorf(helmCommon.CouldNotLoadRepositoriesFileErrorFormat, s.helmhome.RepositoryFile())
		}

		return nil, err
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := s.helmhome.CacheIndex(n)
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			fmt.Fprintf(s.out, "WARNING: Repo %q is corrupt or missing. Try 'helm repo update'.\n", n)
			continue
		}

		i.AddRepo(n, ind, s.versions || len(s.version) > 0)
	}
	return i, nil
}
