package prune

type Options struct{}

type Report struct {
	ItemsDeleted   []string
	SpaceReclaimed uint64
}
