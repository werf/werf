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

// WithWorker annotates the log line with the parallel worker that processed
// it. Under a dynamic scheduler a single worker can process many images
// back to back, and the terminal printer drains one worker's output in
// full before moving to the next (see parallel.Printer) — so a jump in the
// progress index between two consecutive lines is expected whenever the
// worker number also changes, and this annotation makes that visible
// instead of just looking like a scrambled sequence.
func WithWorker(id int) Option {
	return func(o *Options) {
		o.hasWorker = true
		o.worker = id
	}
}

type Options struct {
	withProgress bool
	index        int
	total        int

	hasWorker bool
	worker    int
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

	if options.hasWorker {
		res += fmt.Sprintf(" (worker %d)", options.worker)
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
