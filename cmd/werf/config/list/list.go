package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/werf/werf/cmd/werf/common"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/werf"
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
			if err := common.ProcessLogOptions(&commonCmdData); err != nil {
				common.PrintHelp(cmd)
				return err
			}

			return run()
		},
	}

	cmd.Flags().BoolVarP(&cmdData.imagesOnly, "images-only", "", false, "Show image names without artifacts")

	common.SetupDir(&commonCmdData, cmd)
	common.SetupConfigPath(&commonCmdData, cmd)
	common.SetupConfigTemplatesDir(&commonCmdData, cmd)
	common.SetupTmpDir(&commonCmdData, cmd)
	common.SetupHomeDir(&commonCmdData, cmd)

	common.SetupLogOptions(&commonCmdData, cmd)

	return cmd
}

func run() error {
	if err := werf.Init(*commonCmdData.TmpDir, *commonCmdData.HomeDir); err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}

	tmp_manager.AutoGCEnabled = false

	projectDir, err := common.GetProjectDir(&commonCmdData)
	if err != nil {
		return fmt.Errorf("getting project dir failed: %s", err)
	}

	werfConfig, err := common.GetRequiredWerfConfig(projectDir, &commonCmdData, false)
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
