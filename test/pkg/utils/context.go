package utils

import (
	"context"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/docker"
	"github.com/werf/werf/v2/pkg/logging"
)

func WithDependencies(ctx context.Context) context.Context {
	ctx = logging.WithLogger(ctx)

	if !docker.IsContext(ctx) {
		var err error
		ctx, err = docker.NewContext(ctx)
		Expect(err).ShouldNot(HaveOccurred())
	}

	return ctx
}
