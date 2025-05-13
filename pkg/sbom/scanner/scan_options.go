package scanner

type ScanOptions struct {
	Image      string
	PullPolicy PullPolicy
	Commands   []ScanCommand
}
