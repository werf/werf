package helm

import (
	"context"
	"io"

	"k8s.io/helm/pkg/proto/hapi/services"
)

type GetOptions struct {
	Revision int32
	Template string
}

func Get(out io.Writer, releaseName string, opts GetOptions) error {
	res, err := tillerReleaseServer.GetReleaseContent(context.Background(), &services.GetReleaseContentRequest{
		Name:    releaseName,
		Version: opts.Revision,
	})
	if err != nil {
		return err
	}
	if opts.Template != "" {
		return tpl(opts.Template, res, out)
	}
	return printRelease(out, res.Release)
}
