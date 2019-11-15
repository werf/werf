package logging

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/flant/logboek"
	"github.com/mattn/go-isatty"
)

var (
	imageNameFormat    = "⛵ image %s"
	artifactNameFormat = "🛸 artifact %s"
)

func Init() error {
	if err := logboek.Init(); err != nil {
		return err
	}

	if runtime.GOOS == "windows" && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		DisableLogColor()
	}

	logboek.EnableFitMode()

	log.SetOutput(logboek.GetOutStream())

	return nil
}

func Mute() {
	logboek.MuteOut()
	logboek.MuteErr()
	log.SetOutput(logboek.GetOutStream())
}

func EnableLogColor() {
	logboek.EnableLogColor()
}

func DisableLogColor() {
	logboek.DisableLogColor()
}

func SetWidth(value int) {
	logboek.SetWidth(value)
}

func DisablePrettyLog() {
	imageNameFormat = "image %s"
	artifactNameFormat = "artifact %s"

	logboek.DisablePrettyLog()
}

func ImageLogName(name string, isArtifact bool) string {
	if !isArtifact {
		if name == "" {
			name = "~"
		}
	}

	return name
}

func ImageLogProcessName(name string, isArtifact bool) string {
	logName := ImageLogName(name, isArtifact)
	if !isArtifact {
		return fmt.Sprintf(imageNameFormat, logName)
	} else {
		return fmt.Sprintf(artifactNameFormat, logName)
	}
}
