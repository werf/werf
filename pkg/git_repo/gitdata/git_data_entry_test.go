package gitdata

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func archiveEntry(name string, ago time.Duration) GitDataEntry {
	return &GitArchiveDesc{
		MetadataPath: name,
		ArchivePath:  name + ".tar",
		Metadata:     &ArchiveMetadata{LastAccessTimestamp: time.Now().Add(-ago).Unix()},
	}
}

func patchEntry(name string, ago time.Duration) GitDataEntry {
	return &GitPatchDesc{
		MetadataPath: name,
		PatchPath:    name + ".patch",
		Metadata:     &PatchMetadata{LastAccessTimestamp: time.Now().Add(-ago).Unix()},
	}
}

func worktreeEntry(name string, ago time.Duration, hasSubmodules bool) GitDataEntry {
	return &GitWorktreeDesc{
		Path:          name,
		LastAccessAt:  time.Now().Add(-ago),
		HasSubmodules: hasSubmodules,
	}
}

func mirrorEntry(name string, ago time.Duration) GitDataEntry {
	return &GitRepoDesc{Path: name, LastAccessAt: time.Now().Add(-ago)}
}

func entryNames(entries []GitDataEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.GetPaths()[0])
	}
	return names
}

var _ = Describe("keepGitDataByLru", func() {
	It("removes by priority band first and by age inside a band, against plain LRU order", func() {
		// Ages are interleaved so that plain LRU would give the exact
		// reverse of the expected order for the first and last entries.
		entries := keepGitDataByLru([]GitDataEntry{
			mirrorEntry("mirror-old", 100*time.Hour),
			mirrorEntry("mirror-older", 200*time.Hour),
			worktreeEntry("submodules-old", 50*time.Hour, true),
			worktreeEntry("submodules-older", 60*time.Hour, true),
			worktreeEntry("worktree-old", 20*time.Hour, false),
			worktreeEntry("worktree-older", 30*time.Hour, false),
			archiveEntry("archive", 4*time.Hour),
			patchEntry("patch", 5*time.Hour),
		})

		Expect(entryNames(entries)).To(Equal([]string{
			"patch", "archive",
			"worktree-older", "worktree-old",
			"submodules-older", "submodules-old",
			"mirror-older", "mirror-old",
		}))
	})

	It("never removes entries accessed within the preservation window", func() {
		entries := keepGitDataByLru([]GitDataEntry{
			archiveEntry("archive-fresh", time.Hour),
			worktreeEntry("worktree-fresh", 2*time.Hour+59*time.Minute, false),
			mirrorEntry("mirror-fresh", time.Minute),
			mirrorEntry("mirror-stale", 3*time.Hour+time.Minute),
		})

		Expect(entryNames(entries)).To(Equal([]string{"mirror-stale"}))
	})

	It("orders same-band entries of equal age by path", func() {
		sameAge := time.Now().Add(-10 * time.Hour)
		entries := keepGitDataByLru([]GitDataEntry{
			&GitRepoDesc{Path: "bbb", LastAccessAt: sameAge},
			&GitRepoDesc{Path: "aaa", LastAccessAt: sameAge},
		})

		Expect(entryNames(entries)).To(Equal([]string{"aaa", "bbb"}))
	})

	It("ranks shallow and full mirrors equally, so only age separates them", func() {
		shallow := mirrorEntry("shallow", 11*time.Hour)
		full := mirrorEntry("full", 12*time.Hour)

		Expect(getGitDataEntryRank(shallow)).To(Equal(getGitDataEntryRank(full)))
		Expect(entryNames(keepGitDataByLru([]GitDataEntry{shallow, full}))).To(Equal([]string{"full", "shallow"}))
	})
})
