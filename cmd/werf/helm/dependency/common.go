package dependency

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/util"
)

func isNoRepositoryDefinitionError(err error) bool {
	return strings.HasPrefix(err.Error(), "no repository definition for")
}

func processNoRepositoryDefinitionError(err error) error {
	return fmt.Errorf(strings.Replace(err.Error(), "helm repo add", "werf helm repo add", -1))
}

func getWerfChartPath(commonCmdData common.CmdData) (string, error) {
	var projectDirOrChartDir string

	projectDirOrChartDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if *commonCmdData.Dir != "" {
		if filepath.IsAbs(*commonCmdData.Dir) {
			projectDirOrChartDir = *commonCmdData.Dir
		} else {
			projectDirOrChartDir = filepath.Clean(filepath.Join(projectDirOrChartDir, *commonCmdData.Dir))
		}
	}

	chartDir := filepath.Join(projectDirOrChartDir, ".helm")
	exist, err := util.DirExists(chartDir)
	if err != nil {
		return "", err
	}

	if exist {
		return chartDir, nil
	} else {
		return projectDirOrChartDir, nil
	}
}
