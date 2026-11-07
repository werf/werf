package bundles

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/werfdata"
	"github.com/werf/werf/v3/pkg/logging"
)

var _ = Describe("ResolveRelease", func() {
	packRecordedRelease := func(ctx context.Context, publishEnv string) string {
		GinkgoHelper()

		packOpts := action.ChartPackOptions{
			Destination: GinkgoT().TempDir(),
			Release:     &werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns"},
		}
		if publishEnv != "" {
			packOpts.ExtraValues = map[string]any{"werf": map[string]any{"env": publishEnv}}
		}

		packed, err := action.ChartPack(ctx, writeTestChart(), packOpts)
		Expect(err).To(Succeed())

		return packed.ArchivePath
	}

	DescribeTable("should use the recorded release and namespace",
		func(ctx0 context.Context, release, namespace, env, expectedRelease, expectedNamespace string) {
			ctx := logging.WithLogger(ctx0)

			resolvedRelease, resolvedNamespace, err := ResolveRelease(ctx, packRecordedRelease(ctx, "production"), release, namespace, env)
			Expect(err).To(Succeed())
			Expect(resolvedRelease).To(Equal(expectedRelease))
			Expect(resolvedNamespace).To(Equal(expectedNamespace))
		},
		Entry("without env", "", "", "", "app-production", "app-ns"),
		Entry("with the publish env", "", "", "production", "app-production", "app-ns"),
		Entry("with another env and both flags", "release", "namespace", "staging", "release", "namespace"),
	)

	DescribeTable("should name the flags required for another env",
		func(ctx0 context.Context, publishEnv, release, namespace, expectedErr string) {
			ctx := logging.WithLogger(ctx0)

			_, _, err := ResolveRelease(ctx, packRecordedRelease(ctx, publishEnv), release, namespace, "staging")
			Expect(err).To(MatchError(expectedErr))
		},
		Entry("no flags", "production", "", "",
			`release defaults recorded for env "production" are not used with --env "staging": pass --release and --namespace`),
		Entry("only the release", "production", "release", "",
			`release defaults recorded for env "production" are not used with --env "staging": pass --namespace`),
		Entry("only the namespace", "production", "", "namespace",
			`release defaults recorded for env "production" are not used with --env "staging": pass --release`),
		Entry("published without env", "", "", "",
			`release defaults recorded for env "" are not used with --env "staging": pass --release and --namespace`),
	)

	It("should name the flags required for another env in an unpacked bundle", func(ctx0 context.Context) {
		ctx := logging.WithLogger(ctx0)

		chartDir := writeTestChart()
		files, err := (&werfdata.Data{
			Metadata:    werfdata.Metadata{Release: &werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns"}},
			ExtraValues: map[string]any{"werf": map[string]any{"env": "production"}},
		}).ChartFiles()
		Expect(err).To(Succeed())
		for _, f := range files {
			path := filepath.Join(chartDir, filepath.FromSlash(f.Name))
			Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
			Expect(os.WriteFile(path, f.Data, 0o644)).To(Succeed())
		}

		_, _, err = ResolveRelease(ctx, chartDir, "", "", "staging")
		Expect(err).To(MatchError(`release defaults recorded for env "production" are not used with --env "staging": pass --release and --namespace`))
	})
})
