package bsign

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/werf/exec"
)

// Verify is a wrapper around bsign --verify
// https://manpages.debian.org/testing/bsign/bsign.1.en.html
func Verify(ctx context.Context, _ []string, filename string) error {
	cmd := exec.CommandContextCancellation(ctx, "bsign", "--verify", filename)
	output, err := cmd.CombinedOutput()

	logboek.Context(ctx).Debug().LogF("bsign verify %s output: %s", filename, string(output))

	return err
}
