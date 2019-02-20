package logging

import (
	"fmt"

	"github.com/flant/werf/pkg/logger"
)

var (
	imageNameFormat    = "⛵ image %s"
	artifactNameFormat = "🛸 artifact %s"
)

func Init() {
	logger.Init()
}

func DisablePrettyLog() {
	imageNameFormat = "image %s"
	artifactNameFormat = "artifact %s"

	logger.DisablePrettyLog()
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
