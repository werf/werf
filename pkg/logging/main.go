package logging

import (
	"fmt"

	"github.com/gookit/color"
)

var (
	finalImageNameFormat        = "🛳️  image %s"
	intermediateImageNameFormat = "🏗️️  image %s"
)

func ImageLogName(name string) string {
	if name == "" {
		name = "~"
	}
	return name
}

func ImageLogProcessName(name string, isFinal bool, targetPlatform string) string {
	appendPlatformFunc := func(name string) string {
		if targetPlatform == "" {
			return name
		}
		return fmt.Sprintf("%s [%s]", name, targetPlatform)
	}

	logName := ImageLogName(name)

	if isFinal {
		return appendPlatformFunc(fmt.Sprintf(finalImageNameFormat, logName))
	} else {
		return appendPlatformFunc(fmt.Sprintf(intermediateImageNameFormat, logName))
	}
}

func DisablePrettyLog() {
	finalImageNameFormat = "image %s"
	intermediateImageNameFormat = "image %s"
}

func ImageDefaultStyle(isFinal bool) color.Style {
	var colors []color.Color
	if isFinal {
		colors = []color.Color{color.FgYellow, color.Bold}
	} else {
		colors = []color.Color{color.FgCyan, color.Bold}
	}

	return color.New(colors...)
}

func ImageMetadataStyle() color.Style {
	return ImageDefaultStyle(false)
}
