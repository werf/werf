package gitdata

import "time"

type GitDataEntry interface {
	GetPaths() []string
	GetSize() uint64
	GetLastAccessAt() time.Time
	GetCacheBasePath() string
}

type GitDataLruSort []GitDataEntry

func (a GitDataLruSort) Len() int { return len(a) }
func (a GitDataLruSort) Less(i, j int) bool {
	return a[i].GetLastAccessAt().Before(a[j].GetLastAccessAt())
}
func (a GitDataLruSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func PreserveGitDataByLru(entries []GitDataEntry) []GitDataEntry {
	var res []GitDataEntry

	for _, entry := range entries {
		if !ShouldPreserveGitDataEntryByLru(entry) {
			res = append(res, entry)
		}
	}

	return res
}

func ShouldPreserveGitDataEntryByLru(entry GitDataEntry) bool {
	return time.Since(entry.GetLastAccessAt()) < 3*time.Hour
}
