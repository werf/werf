package true_git

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/path_matcher"
)

var _ = ginkgo.Describe("content-based archive identity", func() {
	ginkgo.DescribeTable("does not reuse keys written before checkout-input fingerprints",
		func(namespace string) {
			opts := ArchiveOptions{ContentChecksum: "content-a", PathScope: "app"}
			gomega.Expect(opts.ID()).NotTo(gomega.Equal(util.Sha256Hash(namespace, opts.ContentChecksum, opts.PathScope, opts.Owner, opts.Group)))
		},
		ginkgo.Entry("unguarded keys", "dockerfile-context-v1"),
		ginkgo.Entry("guarded content-only keys", "dockerfile-context-v2"),
	)

	ginkgo.It("keeps the legacy identity when content checksums are not requested", func() {
		opts := ArchiveOptions{Commit: "commit-a", PathScope: "app", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{})}
		gomega.Expect(opts.ID()).To(gomega.Equal(util.Sha256Hash(opts.Commit, opts.PathScope, opts.PathMatcher.ID())))
		opts.FileRenames = map[string]string{"app/a": "b"}
		gomega.Expect(opts.ID()).To(gomega.Equal(util.Sha256Hash("app/a", "b", opts.Commit, opts.PathScope, opts.PathMatcher.ID())))
	})

	ginkgo.DescribeTable("keys content archives independently of commit and filter identity",
		func(change func(*ArchiveOptions), same bool) {
			opts := ArchiveOptions{ContentChecksum: "content-a", Commit: "commit-a", PathScope: "app", PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}), FileRenames: map[string]string{"app/source": "before"}}
			id := opts.ID()
			change(&opts)
			gomega.Expect(opts.ID() == id).To(gomega.Equal(same))
		},
		ginkgo.Entry("another commit", func(opts *ArchiveOptions) { opts.Commit = "commit-b" }, true),
		ginkgo.Entry("another filter selecting the same content", func(opts *ArchiveOptions) {
			opts.PathMatcher = path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{BasePath: "app"})
		}, true),
		ginkgo.Entry("changed content", func(opts *ArchiveOptions) { opts.ContentChecksum = "content-b" }, false),
		ginkgo.Entry("legacy namespace", func(opts *ArchiveOptions) { opts.ContentChecksum = "" }, false),
		ginkgo.Entry("another scope", func(opts *ArchiveOptions) { opts.PathScope = "other" }, false),
		ginkgo.Entry("another owner", func(opts *ArchiveOptions) { opts.Owner = "1000" }, false),
		ginkgo.Entry("another group", func(opts *ArchiveOptions) { opts.Group = "1000" }, false),
		ginkgo.Entry("renamed file", func(opts *ArchiveOptions) { opts.FileRenames = map[string]string{"app/a": "b"} }, false),
		ginkgo.Entry("changed rename destination", func(opts *ArchiveOptions) { opts.FileRenames["app/source"] = "after" }, false),
	)
})
