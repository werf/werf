package logging

import (
	"fmt"
	"log"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/logboek/pkg/style"
)

var (
	imageNameFormat    = "⛵ image %s"
	artifactNameFormat = "🛸 artifact %s"
)

func Init() error {
	logboek.Streams().EnableLineWrapping()
	log.SetOutput(logboek.ProxyOutStream())
	return nil
}

func EnableLogQuiet() {
	logboek.Streams().Mute()
}

func EnableLogDebug() {
	logboek.SetAcceptedLevel(level.Debug)
	logboek.Streams().SetPrefixStyle(style.Details())
}

func EnableLogVerbose() {
	logboek.SetAcceptedLevel(level.Info)
}

func EnableLogColor() {
	logboek.Streams().EnableStyle()
}

func DisableLogColor() {
	logboek.Streams().DisableStyle()
}

func SetWidth(value int) {
	logboek.Streams().SetWidth(value)
}

func DisablePrettyLog() {
	imageNameFormat = "image %s"
	artifactNameFormat = "artifact %s"

	logboek.Streams().DisablePrettyLog()
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
