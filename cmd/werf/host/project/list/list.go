package list

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-units"

	"github.com/flant/logboek"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/docker"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/werf"
)

var cmdData struct {
	ModifiedBeforeInSeconds int64
	ShowNamesOnly           bool
}

var commonCmdData common.CmdData

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "list",
		Short:                 "List project names based on local stages storage",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run()
		},
	}

	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	common.SetupDockerConfig(&commonCmdData, cmd, "")

	cmd.Flags().Int64VarP(&cmdData.ModifiedBeforeInSeconds, "modified-before", "", -1, "Print project names that have been modified before the timestamp")
	if err := cmd.Flags().MarkHidden("modified-before"); err != nil {
		panic(err)
	}

	cmd.Flags().BoolVarP(&cmdData.ShowNamesOnly, "names-only", "q", false, "Only show project names")

	return cmd
}

func run() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	if err := docker.Init(*commonCmdData.DockerConfig, *commonCmdData.LogVerbose, *commonCmdData.LogDebug); err != nil {
		return err
	}

	projects, err := getProjects()
	if err != nil {
		return err
	}

	projects, err = filterProjects(projects)
	if err != nil {
		return err
	}

	printProjects(projects)

	return nil
}

type projectFields struct {
	Created  int64
	Modified int64
}

func getProjects() (map[string]*projectFields, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("label", imagePkg.WerfLabel)
	options := types.ImageListOptions{Filters: filterSet}

	images, err := docker.Images(options)
	if err != nil {
		return nil, err
	}

	projects := map[string]*projectFields{}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			repo := strings.Split(tag, ":")[0]
			if strings.HasPrefix(repo, storage.LocalStage_ImageRepoPrefix) {
				projectName := strings.TrimPrefix(repo, storage.LocalStage_ImageRepoPrefix)
				project, exist := projects[projectName]
				if !exist {
					projects[projectName] = &projectFields{
						Created:  image.Created,
						Modified: image.Created,
					}
				} else {
					if image.Created < project.Created {
						project.Created = image.Created
					}

					if image.Created > project.Modified {
						project.Modified = image.Created
					}
				}
			}
		}
	}

	return projects, nil
}

func filterProjects(projects map[string]*projectFields) (map[string]*projectFields, error) {
	if cmdData.ModifiedBeforeInSeconds == -1 {
		return projects, nil
	}

	modifiedBefore := time.Now().UTC().Truncate(time.Duration(cmdData.ModifiedBeforeInSeconds) * time.Second)
	newProjects := map[string]*projectFields{}
	for projectName, project := range projects {
		if time.Unix(project.Modified, 0).Before(modifiedBefore) {
			newProjects[projectName] = project
		}
	}

	return newProjects, nil
}

func printProjects(projects map[string]*projectFields) {
	if cmdData.ShowNamesOnly {
		for projectName := range projects {
			fmt.Println(projectName)
		}
	} else {
		t := uitable.New()
		t.MaxColWidth = uint(logboek.ContentWidth())
		t.AddRow("NAME", "CREATED", "MODIFIED")
		for projectName, project := range projects {
			now := time.Now().UTC()
			created := units.HumanDuration(now.Sub(time.Unix(project.Created, 0))) + " ago"
			modified := units.HumanDuration(now.Sub(time.Unix(project.Modified, 0))) + " ago"
			t.AddRow(projectName, created, modified)
		}
		fmt.Println(t.String())
	}
}
