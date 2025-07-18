package scanner

import "github.com/werf/common-go/pkg/util"

type ScanOptions struct {
	Image      string
	PullPolicy PullPolicy
	Commands   []ScanCommand
}

func (o ScanOptions) Checksum() string {
	args := make([]string, 0, 1+len(o.Commands))

	args = append(args, o.Image)

	for _, scanCmd := range o.Commands {
		args = append(args, scanCmd.Checksum())
	}

	return util.Sha256Hash(args...)
}

func DefaultSyftScanOptions() ScanOptions {
	return ScanOptions{
		Image:      "ghcr.io/anchore/syft:v1.23.1",
		PullPolicy: PullIfMissing,
		Commands: []ScanCommand{
			NewSyftScanCommand(),
		},
	}
}
