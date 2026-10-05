package gitdata

import (
	"cmp"
	"slices"
	"time"

	"github.com/samber/lo"
)

type GitDataEntry interface {
	GetPaths() []string
	GetSize() uint64
	GetLastAccessAt() time.Time
	GetCacheBasePath() string
}

type gitDataEntryRank int

const (
	gitDataEntryRankArchiveOrPatch         gitDataEntryRank = 0
	gitDataEntryRankWorktree               gitDataEntryRank = 1
	gitDataEntryRankWorktreeWithSubmodules gitDataEntryRank = 2
	gitDataEntryRankMirror                 gitDataEntryRank = 3
)

func getGitDataEntryRank(entry GitDataEntry) gitDataEntryRank {
	switch desc := entry.(type) {
	case *GitArchiveDesc, *GitPatchDesc:
		return gitDataEntryRankArchiveOrPatch
	case *GitWorktreeDesc:
		if desc.HasSubmodules {
			return gitDataEntryRankWorktreeWithSubmodules
		}
		return gitDataEntryRankWorktree
	default:
		return gitDataEntryRankMirror
	}
}

func shouldPreserveGitDataEntryByLru(entry GitDataEntry) bool {
	return time.Since(entry.GetLastAccessAt()) < 3*time.Hour
}

// keepGitDataByLru filters and sorts GitDataEntry entries based on the LRU.
func keepGitDataByLru(entries []GitDataEntry) []GitDataEntry {
	filteredEntries := lo.Filter(entries, func(entry GitDataEntry, _ int) bool {
		return !shouldPreserveGitDataEntryByLru(entry)
	})

	slices.SortFunc(filteredEntries, func(a, b GitDataEntry) int {
		return cmp.Or(
			cmp.Compare(getGitDataEntryRank(a), getGitDataEntryRank(b)),
			a.GetLastAccessAt().Compare(b.GetLastAccessAt()),
			cmp.Compare(getFirstGitDataEntryPath(a), getFirstGitDataEntryPath(b)),
		)
	})

	return filteredEntries
}

func getFirstGitDataEntryPath(entry GitDataEntry) string {
	paths := entry.GetPaths()
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
