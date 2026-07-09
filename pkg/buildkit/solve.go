package buildkit

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	gwclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"

	"github.com/werf/logboek"
)

type SolveOptions struct {
	Repo        string
	ImageConfig []byte
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

	statusCh := make(chan *client.SolveStatus)
	eg, egCtx := errgroup.WithContext(ctx)

	var resp *client.SolveResponse
	eg.Go(func() error {
		var err error
		resp, err = cl.Build(egCtx, solveOpt, "werf", func(ctx context.Context, c gwclient.Client) (*gwclient.Result, error) {
			r, err := c.Solve(ctx, gwclient.SolveRequest{Definition: def.ToPB(), Evaluate: true})
			if err != nil {
				return nil, err
			}

			ref, err := r.SingleRef()
			if err != nil {
				return nil, err
			}

			res := gwclient.NewResult()
			res.SetRef(ref)
			if opts.ImageConfig != nil {
				res.AddMeta(exptypes.ExporterImageConfigKey, opts.ImageConfig)
			}
			return res, nil
		}, statusCh)
		if err != nil {
			return fmt.Errorf("solve: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		display, err := progressui.NewDisplay(logboek.Context(ctx).OutStream(), progressui.PlainMode)
		if err != nil {
			return fmt.Errorf("create progress display: %w", err)
		}
		if _, err := display.UpdateFrom(context.WithoutCancel(egCtx), statusCh); err != nil {
			return fmt.Errorf("display build progress: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", err
	}

	imageDigest := resp.ExporterResponse[exptypes.ExporterImageDigestKey]
	if imageDigest == "" {
		return "", fmt.Errorf("no image digest in solve response")
	}

	return fmt.Sprintf("%s@%s", opts.Repo, imageDigest), nil
}
