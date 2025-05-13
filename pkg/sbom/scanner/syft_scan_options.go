package scanner

func DefaultSyftScanOptions() *ScanOptions {
	return &ScanOptions{
		Image:      "ghcr.io/anchore/syft:v1.23.1",
		PullPolicy: PullIfMissing,
		Commands: []ScanCommand{
			NewSyftScanCommand(),
		},
	}
}
