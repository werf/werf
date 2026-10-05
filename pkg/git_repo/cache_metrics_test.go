package git_repo

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/opstats"
	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/pkg/true_git"
	"github.com/werf/werf/v3/pkg/werf"
	"github.com/werf/werf/v3/test/pkg/utils"
)

var _ = ginkgo.Describe("Git cache lookup counters", func() {
	var collector *opstats.Collector

	rows := func(ctx ginkgo.SpecContext) map[string]opstats.CacheSummary {
		res := make(map[string]opstats.CacheSummary)
		for _, summary := range collector.CacheSummary(ctx) {
			res[string(summary.Operation)+"/"+string(summary.Layer)] = summary
		}
		return res
	}

	operationCount := func(op opstats.Operation) int {
		for _, summary := range collector.Summary() {
			if summary.Operation == op {
				return summary.Count
			}
		}
		return 0
	}

	taggedRepo := func(ctx ginkgo.SpecContext) string {
		source := newChecksumTestRepo(ctx)
		gomega.Expect(os.WriteFile(filepath.Join(source, "file"), []byte("data"), 0o644)).To(gomega.Succeed())
		commitChecksumTestRepo(ctx, source)
		utils.RunSucceedCommand(ctx, source, "git", "-c", "tag.gpgsign=false", "tag", "v1")
		return source
	}

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
		gomega.Expect(true_git.Init(ctx, true_git.Options{})).To(gomega.Succeed())
		gomega.Expect(Init(&fakeGitDataManager{})).To(gomega.Succeed())
		resetLsRemoteTagsCache()
		collector = opstats.NewCollector()
	})

	ginkgo.It("counts remote tag listings on the memory layer only", func(ctx ginkgo.SpecContext) {
		source := taggedRepo(ctx)
		countedCtx := opstats.NewContext(ctx, collector)
		repo := &Remote{Url: source, Tag: "v1"}

		cold, err := repo.lsRemoteTag(countedCtx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(cold).NotTo(gomega.BeEmpty())

		warm, err := repo.lsRemoteTag(countedCtx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(warm).To(gomega.Equal(cold))

		fresh, err := repo.lsRemoteTag(countedCtx, true)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(fresh).To(gomega.Equal(cold))

		// A ready listing that does not contain the requested tag answers from the
		// cache: the user error is not a cache miss.
		missing, err := OpenRemoteRepo("absent-tag", source, nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		missing.Tag = "absent"
		_, err = missing.lsRemoteTag(countedCtx, false)
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(`bad tag "absent"`)))

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: ls-remote/memory": {
				Operation: opstats.OperationGitLsRemote,
				Layer:     opstats.CacheLayerMemory,
				Hit:       2,
				Miss:      1,
				Bypass:    1,
			},
		}))
		// The listing ran for the miss and for the explicit fresh lookup, and for
		// neither hit.
		gomega.Expect(operationCount(opstats.OperationGitLsRemote)).To(gomega.Equal(2))
	})

	ginkgo.It("counts a fresh listing as a bypass even with nothing cached", func(ctx ginkgo.SpecContext) {
		source := taggedRepo(ctx)
		countedCtx := opstats.NewContext(ctx, collector)
		repo := &Remote{Url: source, Tag: "v1"}

		sha, err := repo.lsRemoteTag(countedCtx, true)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(sha).NotTo(gomega.BeEmpty())

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: ls-remote/memory": {
				Operation: opstats.OperationGitLsRemote,
				Layer:     opstats.CacheLayerMemory,
				Bypass:    1,
			},
		}))
	})

	ginkgo.It("counts checksums on the memory layer and runs the computation once", func(ctx ginkgo.SpecContext) {
		source := newChecksumTestRepo(ctx)
		gomega.Expect(os.WriteFile(filepath.Join(source, "file"), []byte("data"), 0o644)).To(gomega.Succeed())
		commitChecksumTestRepo(ctx, source)

		remote, err := OpenRemoteRepo("checksums", source, nil)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remote.CloneAndFetch(ctx)).To(gomega.Succeed())

		countedCtx := opstats.NewContext(ctx, collector)
		opts := ChecksumOptions{
			Commit:        utils.GetHeadCommit(ctx, source),
			LsTreeOptions: LsTreeOptions{PathMatcher: path_matcher.NewTruePathMatcher(), AllFiles: true},
		}

		first, err := remote.GetOrCreateChecksum(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: checksum/memory": {Operation: opstats.OperationGitChecksum, Layer: opstats.CacheLayerMemory, Miss: 1},
		}))
		second, err := remote.GetOrCreateChecksum(countedCtx, opts)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(second).To(gomega.Equal(first))

		gomega.Expect(rows(ctx)).To(gomega.Equal(map[string]opstats.CacheSummary{
			"git: checksum/memory": {
				Operation: opstats.OperationGitChecksum,
				Layer:     opstats.CacheLayerMemory,
				Hit:       1,
				Miss:      1,
			},
		}))
		gomega.Expect(operationCount(opstats.OperationGitChecksum)).To(gomega.Equal(1))
	})

	ginkgo.It("records nothing when no collector is bound", func(ctx ginkgo.SpecContext) {
		repo := &Remote{Url: taggedRepo(ctx), Tag: "v1"}
		_, err := repo.lsRemoteTag(ctx, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(collector.CacheSummary(ctx)).To(gomega.BeEmpty())
	})
})
