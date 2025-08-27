package gitdata

import (
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

func shouldPreserveGitDataEntryByLru(entry GitDataEntry) bool {
	return time.Since(entry.GetLastAccessAt()) < 3*time.Hour
}

// keepGitDataByLru filters and sorts GitDataEntry entries based on the LRU.
func keepGitDataByLru(entries []GitDataEntry) []GitDataEntry {
	filteredEntries := lo.Filter(entries, func(entry GitDataEntry, _ int) bool {
		return !shouldPreserveGitDataEntryByLru(entry)
	})

	slices.SortFunc(filteredEntries, func(a, b GitDataEntry) int {
		return a.GetLastAccessAt().Compare(b.GetLastAccessAt())
	})

	return filteredEntries
}
