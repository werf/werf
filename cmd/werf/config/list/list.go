package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/werf"
)

var commonCmdData common.CmdData
var cmdData struct {
	imagesOnly bool
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "list",
		DisableFlagsInUseLine: true,
		Short:                 "List image and artifact names defined in werf.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run()
		},
	}

	cmd.Flags().BoolVarP(&cmdData.imagesOnly, "images-only", "", false, "Show image names without artifacts")

	common.SetupDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	return cmd
}

func run() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfigPath, err := common.GetWerfConfigPath(projectDir)
	if err != nil {
		return err
	}

	werfConfig, err := config.GetWerfConfig(werfConfigPath, false)
	if err != nil {
		return err
	}

	for _, image := range werfConfig.StapelImages {
		if image.Name == "" {
			fmt.Println("~")
		} else {
			fmt.Println(image.Name)
		}
	}

	for _, image := range werfConfig.ImagesFromDockerfile {
		if image.Name == "" {
			fmt.Println("~")
		} else {
			fmt.Println(image.Name)
		}
	}

	if !cmdData.imagesOnly {
		for _, image := range werfConfig.Artifacts {
			fmt.Println(image.Name)
		}
	}

	return nil
}
