package buildkit

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/session"
	"github.com/tonistiigi/fsutil"
)

type SolveOptions struct {
	Repo        string
	LocalMounts map[string]fsutil.FS
	Session     []session.Attachable
}

func Solve(ctx context.Context, cl *client.Client, def *llb.Definition, opts SolveOptions) (string, error) {
	if opts.Repo == "" {
		panic("Repo is required for buildkit solve")
	}

	solveOpt := client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterImage,
				Attrs: map[string]string{
					"name":           opts.Repo,
					"push":           "true",
					"push-by-digest": "true",
				},
			},
		},
		LocalMounts: opts.LocalMounts,
		Session:     opts.Session,
	}

	resp, err := cl.Solve(ctx, def, solveOpt, nil)
	if err != nil {
		return "", fmt.Errorf("solve: %w", err)
	}

	imageDigest := resp.ExporterResponse[exptypes.ExporterImageDigestKey]
	if imageDigest == "" {
		return "", fmt.Errorf("no image digest in solve response")
	}

	return fmt.Sprintf("%s@%s", opts.Repo, imageDigest), nil
}
