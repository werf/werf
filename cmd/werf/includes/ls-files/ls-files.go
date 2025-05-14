package lsfiles

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/cmd/werf/common"
	"github.com/werf/werf/v2/pkg/true_git"
)

const (
	tableHeader = "PATH\tSOURCE"
)

var commonCmdData common.CmdData

func NewCmd(ctx context.Context) *cobra.Command {
	ctx = common.NewContextWithCmdData(ctx, &commonCmdData)
	cmd := common.SetCommandContext(ctx, common.SetCommandContext(ctx, &cobra.Command{
		Use:   "ls-files",
		Short: "List files in the project with includes",
		Long:  "List files in the project with includes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			_, ctx, err := common.InitCommonComponents(ctx, common.InitCommonComponentsOptions{
				Cmd:                &commonCmdData,
				InitWerf:           true,
				InitGitDataManager: true,
				InitTrueGitWithOptions: &common.InitTrueGitOptions{
					Options: true_git.Options{LiveGitOutput: *commonCmdData.LogDebug},
				},
			})
			if err != nil {
				return fmt.Errorf("component init error: %w", err)
			}

			gm, err := common.GetGiterminismManager(ctx, &commonCmdData)
			if err != nil {
				return err
			}

			sourceFilters, err := parseSourceFilter(commonCmdData.IncludesLsFilter)
			if err != nil {
				return fmt.Errorf("unable to parse filter: %w", err)
			}

			m, err := gm.FileManager.ListFilesByGlob(ctx, "*", sourceFilters)
			if err != nil {
				return fmt.Errorf("unable to get files: %w", err)
			}

			tb, err := writeTable(m)
			if err != nil {
				return fmt.Errorf("unable to write table: %w", err)
			}

			logboek.Context(ctx).Log(tb)

			return nil
		},
	}))

	common.SetupDir(&commonCmdData, cmd)
	common.SetupGitWorkTree(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupGiterminismConfigPath(&commonCmdData, cmd)
	common.SetupEnvironment(&commonCmdData, cmd)

	common.SetupGiterminismOptions(&commonCmdData, cmd)

	common.SetupTmpDir(&commonCmdData, cmd, common.SetupTmpDirOptions{})
	common.SetupHomeDir(&commonCmdData, cmd, common.SetupHomeDirOptions{})

	common.SetupLogOptions(&commonCmdData, cmd)
	common.SetupLogProjectDir(&commonCmdData, cmd)

	commonCmdData.SetupPlatform(cmd)
	common.SetupFollow(&commonCmdData, cmd)

	commonCmdData.SetupCreateIncludesLockFile(false)
	commonCmdData.SetupUseIncludesLatestVersions(cmd)
	commonCmdData.SetupIncludesLsFilter(cmd)

	return cmd
}

func parseSourceFilter(input *string) ([]string, error) {
	if input == nil {
		return []string{}, nil
	}

	parts := strings.SplitN(*input, "=", 2)
	if len(parts) != 2 || parts[0] != "source" {
		return nil, fmt.Errorf("invalid filter format: must be 'source=...'")
	}

	sources := strings.Split(parts[1], ",")
	for i := range sources {
		sources[i] = strings.TrimSpace(sources[i])
	}
	return sources, nil
}

func parseGlobs() {

}

func writeTable(data map[string]string) (string, error) {
	paths := make([]string, 0, len(data))
	for path := range data {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, tableHeader)
	for _, path := range paths {
		fmt.Fprintf(w, "%s\t%s\n", path, data[path])
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
