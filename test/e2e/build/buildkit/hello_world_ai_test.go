package e2e_buildkit_test

import (
	"fmt"
	"os"

	"github.com/moby/buildkit/client/llb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/buildkit"
)

var _ = Describe("Buildkit hello-world solve", Label("e2e", "buildkit"), func() {
	It("solves a trivial llb.Image graph and pushes by digest", func(ctx SpecContext) {
		host := os.Getenv("WERF_BUILDKIT_HOST")
		if host == "" {
			Skip("WERF_BUILDKIT_HOST is not set")
		}
		registry := os.Getenv("WERF_TEST_K8S_DOCKER_REGISTRY")
		if registry == "" {
			Skip("WERF_TEST_K8S_DOCKER_REGISTRY is not set")
		}
		repo := fmt.Sprintf("%s/buildkit-hello-world", registry)

		cl, err := buildkit.NewClient(ctx, host)
		Expect(err).NotTo(HaveOccurred())
		defer cl.Close()

		def, err := llb.Image("busybox:latest").Marshal(ctx)
		Expect(err).NotTo(HaveOccurred())

		attachables, err := buildkit.SessionAttachables(buildkit.SessionAttachablesOptions{})
		Expect(err).NotTo(HaveOccurred())

		builtID, err := buildkit.Solve(ctx, cl, def, buildkit.SolveOptions{
			Repo:    repo,
			Session: attachables,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(builtID).To(MatchRegexp(fmt.Sprintf(`^%s@sha256:[0-9a-f]{64}$`, repo)))
	})
})
