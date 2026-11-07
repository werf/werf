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
		data, hasWerfData, err := ReadWerfData(ctx, GinkgoT().TempDir())
		Expect(err).To(Succeed())
		Expect(hasWerfData).To(BeFalse())
		Expect(ResolveRelease(ctx, data, ReleaseFlags{})).To(Equal(ReleaseFlags{}))
	})

	It("should read the .werf-data of a bundle with .werf-data", func(ctx context.Context) {
		chartDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(chartDir, werfdata.DirName), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName), []byte("apiVersion: v2\nvalues:\n  extraValues: true\nrelease:\n  name: app-production\n  namespace: app-ns\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, werfdata.DirName, werfdata.ExtraValuesFileName), []byte("werf:\n  env: production\n"), 0o644)).To(Succeed())

		data, hasWerfData, err := ReadWerfData(ctx, chartDir)
		Expect(err).To(Succeed())
		Expect(hasWerfData).To(BeTrue())
		Expect(data.Metadata.Release).To(Equal(&werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns"}))
		Expect(ResolveRelease(ctx, data, ReleaseFlags{Env: "production"})).To(Equal(ReleaseFlags{Release: "app-production", Namespace: "app-ns", Env: "production"}))
	})

	It("should report an unsupported apiVersion before unknown fields", func(ctx context.Context) {
		chartDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(chartDir, werfdata.DirName), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(chartDir, werfdata.DirName, werfdata.MetadataFileName), []byte("apiVersion: v3\nnewField: true\n"), 0o644)).To(Succeed())

		_, _, err := ReadWerfData(ctx, chartDir)
		Expect(err).To(MatchError(ContainSubstring(`unsupported .werf-data/metadata.yaml apiVersion "v3"`)))
	})

	DescribeTable("should resolve the release from the flags and the release defaults",
		func(ctx context.Context, release *werfdata.ReleaseMetadata, flags, expected ReleaseFlags) {
			Expect(ResolveRelease(ctx, recordedWerfData(release, "production"), flags)).To(Equal(expected))
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
		func(ctx context.Context, recordedEnv string, flags ReleaseFlags, expectedErr string) {
			_, err := ResolveRelease(ctx, recordedWerfData(recordedRelease, recordedEnv), flags)
			Expect(err).To(MatchError(expectedErr))
		},
		Entry("no flags", "production", ReleaseFlags{Env: "staging"},
			`release defaults recorded for env "production" are not used with --env "staging": pass --release and --namespace`),
		Entry("only the release", "production", ReleaseFlags{Release: "release", Env: "staging"},
			`release defaults recorded for env "production" are not used with --env "staging": pass --namespace`),
		Entry("only the namespace", "production", ReleaseFlags{Namespace: "namespace", Env: "staging"},
			`release defaults recorded for env "production" are not used with --env "staging": pass --release`),
		Entry("published without env", "", ReleaseFlags{Env: "staging"},
			`release defaults recorded for env "" are not used with --env "staging": pass --release and --namespace`),
	)
})

var recordedRelease = &werfdata.ReleaseMetadata{Name: "app-production", Namespace: "app-ns"}

func recordedWerfData(release *werfdata.ReleaseMetadata, env string) *werfdata.Data {
	data := &werfdata.Data{Metadata: werfdata.Metadata{Release: release}}
	if env != "" {
		data.ExtraValues = map[string]any{"werf": map[string]any{"env": env}}
	}

	return data
}
