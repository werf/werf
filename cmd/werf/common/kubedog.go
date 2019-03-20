package common

import (
	"github.com/flant/kubedog/pkg/display"
	"github.com/flant/logboek"
)

func InitKubedog() error {
	display.SetOut(logboek.GetOutStream())
	display.SetErr(logboek.GetErrStream())

	return nil
}
