package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
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
