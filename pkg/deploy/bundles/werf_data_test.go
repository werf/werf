package bundles

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/nelm/v2/pkg/werfdata"
)

var _ = Describe("Bundle .werf-data", func() {
	It("should read no metadata of a bundle published by werf <= 3.12", func(ctx context.Context) {
		metadata, hasWerfData, err := ReadWerfDataMetadata(ctx, GinkgoT().TempDir())
		Expect(err).To(Succeed())
		Expect(hasWerfData).To(BeFalse())
		Expect(ReleaseDefaults(ctx, metadata, "")).To(Equal(werfdata.ReleaseMetadata{}))
	})

	It("should read the metadata of a bundle with .werf-data", func(ctx context.Context) {
		chartDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(chartDir, werfdata.DirName), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName), []byte("apiVersion: v2\nvalues: {}\nrelease:\n  name: app-production\n  namespace: app-ns\n  env: production\n"), 0o644)).To(Succeed())

		metadata, hasWerfData, err := ReadWerfDataMetadata(ctx, chartDir)
		Expect(err).To(Succeed())
		Expect(hasWerfData).To(BeTrue())
		Expect(metadata.Release).To(Equal(&werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}))
	})

	It("should report an unsupported apiVersion before unknown fields", func(ctx context.Context) {
		chartDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(chartDir, werfdata.DirName), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName), []byte("apiVersion: v3\nnewField: true\n"), 0o644)).To(Succeed())

		_, _, err := ReadWerfDataMetadata(ctx, chartDir)
		Expect(err).To(MatchError(ContainSubstring(`unsupported .werf-data/metadata.yaml apiVersion "v3"`)))
	})

	DescribeTable("should use the release defaults unless the environment differs",
		func(ctx context.Context, release *werfdata.ReleaseMetadata, env string, expected werfdata.ReleaseMetadata) {
			Expect(ReleaseDefaults(ctx, &werfdata.Metadata{Release: release}, env)).To(Equal(expected))
		},
		Entry("no environment given",
			&werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}, "",
			werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}),
		Entry("the same environment",
			&werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}, "production",
			werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}),
		Entry("another environment",
			&werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}, "staging",
			werfdata.ReleaseMetadata{}),
		Entry("no release recorded", nil, "", werfdata.ReleaseMetadata{}),
	)
})
