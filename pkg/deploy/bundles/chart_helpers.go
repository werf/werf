package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/pkg/common"
	"github.com/werf/nelm/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/pkg/helm/pkg/chart/v2/util"
)

func ChartToBytes(ch *chart.Chart) ([]byte, error) {
	chartBytes := bytes.NewBuffer(nil)
	zipper := gzip.NewWriter(chartBytes)
	chartv2util.SetGzipWriterMeta(zipper)
	twriter := tar.NewWriter(zipper)

	if err := chartv2util.SaveIntoTar(twriter, ch, chartv2util.SaveIntoTarOptions{}); err != nil {
		return nil, fmt.Errorf("unable to save chart to tar: %w", err)
	}

	if err := twriter.Close(); err != nil {
		return nil, fmt.Errorf("unable to close chart tar: %w", err)
	}

	if err := zipper.Close(); err != nil {
		return nil, fmt.Errorf("unable to close chart gzip: %w", err)
	}

	return chartBytes.Bytes(), nil
}

func BytesToChart(ctx context.Context, data []byte, opts nelmcommon.HelmOptions) (*chart.Chart, error) {
	dataReader := bytes.NewBuffer(data)

	ch, err := loader.LoadArchive(nelmcommon.ContextWithHelmOptions(ctx, opts), dataReader)
	if err != nil {
		return nil, err
	}

	v2ch, ok := ch.(*chart.Chart)
	if !ok {
		return nil, fmt.Errorf("unsupported chart type %T", ch)
	}

	return v2ch, nil
}

func SaveChartValues(ctx context.Context, ch *chart.Chart) error {
	valuesRaw, err := yaml.Marshal(ch.Values)
	if err != nil {
		return fmt.Errorf("unable to marshal chart values: %w", err)
	}
	logboek.Context(ctx).Debug().LogF("Values after change (%v):\n%s\n---\n", err, valuesRaw)

	for _, f := range ch.Raw {
		if f.Name == chartv2util.ValuesfileName {
			f.Data = valuesRaw
			break
		}
	}

	return nil
}
