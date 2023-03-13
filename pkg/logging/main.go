package logging

import (
	"fmt"
)

var (
	imageNameFormat    = "⛵ image %s"
	artifactNameFormat = "🛸 artifact %s"
)

func ImageLogName(name string, isArtifact bool) string {
	if !isArtifact {
		if name == "" {
			name = "~"
		}
	}

	return name
}

func ImageLogProcessName(name string, isArtifact bool, targetPlatform string) string {
	appendPlatformFunc := func(name string) string {
		if targetPlatform == "" {
			return name
		}
		return fmt.Sprintf("%s [%s]", name, targetPlatform)
	}

	logName := ImageLogName(name, isArtifact)

	if !isArtifact {
		return appendPlatformFunc(fmt.Sprintf(imageNameFormat, logName))
	} else {
		return appendPlatformFunc(fmt.Sprintf(artifactNameFormat, logName))
	}
}

func DisablePrettyLog() {
	imageNameFormat = "image %s"
	artifactNameFormat = "artifact %s"
}
