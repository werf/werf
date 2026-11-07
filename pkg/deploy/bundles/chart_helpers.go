package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader"
	chart "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2"
	chartv2util "github.com/werf/nelm/v2/pkg/helm/pkg/chart/v2/util"
)

// ChartToBytes packs the chart. A chart with .werf-data gets its original Chart.yaml and Chart.lock
// instead of the ones generated from the parsed metadata, so that their formatting survives: keep
// them in sync with the metadata.
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

	if len(ch.WerfDataFiles) == 0 {
		return chartBytes.Bytes(), nil
	}

	replacements := map[string][]byte{}
	for _, f := range ch.Raw {
		if f.Name == chartv2util.ChartfileName || f.Name == "Chart.lock" {
			replacements[path.Join(ch.Name(), f.Name)] = f.Data
		}
	}

	result, err := replaceArchiveFiles(chartBytes.Bytes(), replacements)
	if err != nil {
		return nil, fmt.Errorf("write original Chart.yaml and Chart.lock: %w", err)
	}

	return result, nil
}

func replaceArchiveFiles(archive []byte, replacements map[string][]byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	defer gzipReader.Close()

	result := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(result)
	gzipWriter.Header = gzipReader.Header
	tarReader := tar.NewReader(gzipReader)
	tarWriter := tar.NewWriter(gzipWriter)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}

		body, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read %q from archive: %w", header.Name, err)
		}

		if replacement, ok := replacements[header.Name]; ok {
			body = replacement
			header.Size = int64(len(body))
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("write %q into archive: %w", header.Name, err)
		}

		if _, err := tarWriter.Write(body); err != nil {
			return nil, fmt.Errorf("write %q into archive: %w", header.Name, err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("write archive: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("write archive: %w", err)
	}

	return result.Bytes(), nil
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
