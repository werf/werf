package common

import (
	"github.com/flant/kubedog/pkg/display"
	"github.com/flant/werf/pkg/logger"
)

func InitKubedog() error {
	display.SetOut(logger.GetOutStream())
	display.SetErr(logger.GetErrStream())

	return nil
}
