package gitdata

import "time"

type GitDataEntry interface {
	GetPaths() []string
	GetSize() uint64
	GetLastAccessAt() time.Time
}

type GitDataLruSort []GitDataEntry

func (a GitDataLruSort) Len() int { return len(a) }
func (a GitDataLruSort) Less(i, j int) bool {
	return a[i].GetLastAccessAt().Before(a[j].GetLastAccessAt())
}
func (a GitDataLruSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
