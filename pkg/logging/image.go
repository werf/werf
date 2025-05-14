package logging

import (
	"fmt"

	"github.com/gookit/color"
)

var (
	finalImagePrettyPrefix        = "🛳️  "
	intermediateImagePrettyPrefix = "🏗️️  "
)

func ImageLogName(name string) string {
	if name == "" {
		name = "~"
	}
	return name
}

func WithProgress(index, total int) Option {
	return func(o *Options) {
		o.withProgress = true
		o.index = index
		o.total = total
	}
}

type Options struct {
	withProgress bool
	index        int
	total        int
}

type Option func(*Options)

func ImageLogProcessName(name string, isFinal bool, targetPlatform string, opts ...Option) string {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	var res string
	res += "image" + " " + ImageLogName(name)

	if targetPlatform != "" {
		res += " [" + targetPlatform + "]"
	}

	if options.withProgress {
		res = fmt.Sprintf("(%d/%d)", options.index, options.total) + " " + res
	}

	if isFinal {
		res = finalImagePrettyPrefix + res
	} else {
		res = intermediateImagePrettyPrefix + res
	}

	return res
}

func DisablePrettyLog() {
	finalImagePrettyPrefix = ""
	intermediateImagePrettyPrefix = ""
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
