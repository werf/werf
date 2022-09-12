package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
)

func ChartToBytes(ch *chart.Chart) ([]byte, error) {
	chartBytes := bytes.NewBuffer(nil)
	zipper := gzip.NewWriter(chartBytes)
	chartutil.SetGzipWriterMeta(zipper)
	twriter := tar.NewWriter(zipper)

	if err := chartutil.SaveIntoTar(twriter, ch, chartutil.SaveIntoTarOptions{}); err != nil {
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

func BytesToChart(data []byte) (*chart.Chart, error) {
	dataReader := bytes.NewBuffer(data)
	return loader.LoadArchiveWithOptions(dataReader, loader.LoadOptions{})
}

func SaveChartValues(ctx context.Context, ch *chart.Chart) error {
	valuesRaw, err := yaml.Marshal(ch.Values)
	if err != nil {
		return fmt.Errorf("unable to marshal chart values: %w", err)
	}
	logboek.Context(ctx).Debug().LogF("Values after change (%v):\n%s\n---\n", err, valuesRaw)

	for _, f := range ch.Raw {
		if f.Name == chartutil.ValuesfileName {
			f.Data = valuesRaw
			break
		}
	}

	return nil
}
