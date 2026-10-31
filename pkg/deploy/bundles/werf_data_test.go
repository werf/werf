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
		Expect(ResolveRelease(ctx, metadata, ReleaseFlags{})).To(Equal(ReleaseFlags{}))
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

	DescribeTable("should resolve the release from the flags and the release defaults",
		func(ctx context.Context, release *werfdata.ReleaseMetadata, flags, expected ReleaseFlags) {
			Expect(ResolveRelease(ctx, &werfdata.Metadata{Release: release}, flags)).To(Equal(expected))
		},
		Entry("no flags",
			recordedRelease, ReleaseFlags{},
			ReleaseFlags{Release: "app-production", Namespace: "app-ns"}),
		Entry("the same environment",
			recordedRelease, ReleaseFlags{Env: "production"},
			ReleaseFlags{Release: "app-production", Namespace: "app-ns", Env: "production"}),
		Entry("the flags win",
			recordedRelease, ReleaseFlags{Release: "release", Namespace: "namespace"},
			ReleaseFlags{Release: "release", Namespace: "namespace"}),
		Entry("only the namespace given",
			recordedRelease, ReleaseFlags{Namespace: "namespace"},
			ReleaseFlags{Release: "app-production", Namespace: "namespace"}),
		Entry("another environment with both flags",
			recordedRelease, ReleaseFlags{Release: "release", Namespace: "namespace", Env: "staging"},
			ReleaseFlags{Release: "release", Namespace: "namespace", Env: "staging"}),
		Entry("no release recorded",
			nil, ReleaseFlags{Env: "staging"},
			ReleaseFlags{Env: "staging"}),
	)

	DescribeTable("should require both flags for another environment",
		func(ctx context.Context, flags ReleaseFlags) {
			_, err := ResolveRelease(ctx, &werfdata.Metadata{Release: recordedRelease}, flags)
			Expect(err).To(MatchError(`release defaults recorded for env "production" are not used with --env "staging": pass --release and --namespace`))
		},
		Entry("no flags", ReleaseFlags{Env: "staging"}),
		Entry("only the release", ReleaseFlags{Release: "release", Env: "staging"}),
		Entry("only the namespace", ReleaseFlags{Namespace: "namespace", Env: "staging"}),
	)
})

var recordedRelease = &werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns", Env: "production"}
