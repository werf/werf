package bundles

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"helm.sh/helm/v3/pkg/chart"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

const (
	chartArchiveFileName = "chart.tar.gz"
)

type BundleArchive struct {
	Reader BundleArchiveReader
	Writer BundleArchiveWriter
}

func NewBundleArchive(reader BundleArchiveReader, writer BundleArchiveWriter) *BundleArchive {
	return &BundleArchive{Reader: reader, Writer: writer}
}

func (bundle *BundleArchive) GetImageArchiveOpener(imageTag string) *ImageArchiveOpener {
	return NewImageArchiveOpener(bundle, imageTag)
}

func (bundle *BundleArchive) ReadChart(ctx context.Context) (*chart.Chart, error) {
	chartBytes, err := bundle.Reader.ReadChartArchive()
	if err != nil {
		return nil, fmt.Errorf("unable to read chart archive: %w", err)
	}

	ch, err := BytesToChart(chartBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse chart archive from bundle archive: %w", err)
	}

	return ch, nil
}

func (bundle *BundleArchive) WriteChart(ctx context.Context, ch *chart.Chart) error {
	chartBytes, err := ChartToBytes(ch)
	if err != nil {
		return fmt.Errorf("unable to dump chart to archive: %w", err)
	}

	if err := bundle.Writer.WriteChartArchive(chartBytes); err != nil {
		return fmt.Errorf("unable to write chart archive into bundle archive: %w", err)
	}

	return nil
}

func (bundle *BundleArchive) CopyTo(ctx context.Context, to BundleAccessor, opts copyToOptions) error {
	return to.CopyFromArchive(ctx, bundle, opts)
}

func (bundle *BundleArchive) CopyFromArchive(ctx context.Context, fromArchive *BundleArchive, opts copyToOptions) error {
	ch, err := fromArchive.ReadChart(ctx)
	if err != nil {
		return fmt.Errorf("unable to read chart from the bundle archive %q: %w", fromArchive.Reader.String(), err)
	}

	if err := bundle.Writer.Open(); err != nil {
		return fmt.Errorf("unable to open target bundle archive: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s into archive", fromArchive.Reader.String()).DoError(func() error {
		return bundle.WriteChart(ctx, ch)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).LogProcess("Copy images from bundle archive").DoError(func() error {
		if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
			if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
				for imageName, v := range imageVals {
					if imageRef, ok := v.(string); ok {
						logboek.Context(ctx).Default().LogFDetails("Saving image %s\n", imageRef)

						_, tag := image.ParseRepositoryAndTag(imageRef)

						imageArchive, err := fromArchive.Reader.ReadImageArchive(tag)
						if err != nil {
							return fmt.Errorf("error reading image archive by tag %q from the bundle archive %q: %w", tag, fromArchive.Reader.String(), err)
						}

						imageBytes := bytes.NewBuffer(nil)

						if _, err := io.Copy(imageBytes, imageArchive); err != nil {
							return fmt.Errorf("error reading image archive by tag %q from the bundle archive %q: %w", tag, fromArchive.Reader.String(), err)
						}

						if err := imageArchive.Close(); err != nil {
							return fmt.Errorf("unable to close image archive reader by tag %q from the bundle archive %q: %w", tag, fromArchive.Reader.String(), err)
						}

						if err := bundle.Writer.WriteImageArchive(tag, imageBytes.Bytes()); err != nil {
							return fmt.Errorf("error writing image %q into bundle archive: %w", imageRef, err)
						}
					} else {
						return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
					}
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := bundle.Writer.Save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}

func (bundle *BundleArchive) CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle, opts copyToOptions) error {
	ch, err := fromRemote.ReadChart(ctx)
	if err != nil {
		return fmt.Errorf("unable to read chart from remote bundle: %w", err)
	}

	if err := bundle.Writer.Open(); err != nil {
		return fmt.Errorf("unable to open target bundle archive: %w", err)
	}

	if err := logboek.Context(ctx).LogProcess("Saving bundle %s into archive", fromRemote.RegistryAddress.FullName()).DoError(func() error {
		return bundle.WriteChart(ctx, ch)
	}); err != nil {
		return err
	}

	if werfVals, ok := ch.Values["werf"].(map[string]interface{}); ok {
		if imageVals, ok := werfVals["image"].(map[string]interface{}); ok {
			for imageName, v := range imageVals {
				if imageRef, ok := v.(string); ok {
					logboek.Context(ctx).Default().LogFDetails("Saving image %s\n", imageRef)

					_, tag := image.ParseRepositoryAndTag(imageRef)

					// TODO: maybe save into tmp file archive OR read resulting image size from the registry before pulling
					imageBytes := bytes.NewBuffer(nil)

					if err := fromRemote.RegistryClient.PullImageArchive(ctx, imageBytes, imageRef); err != nil {
						return fmt.Errorf("error pulling image %q archive: %w", imageRef, err)
					}

					if err := bundle.Writer.WriteImageArchive(tag, imageBytes.Bytes()); err != nil {
						return fmt.Errorf("error writing image %q into bundle archive: %w", imageRef, err)
					}
				} else {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}
			}
		}
	}

	if err := bundle.Writer.Save(); err != nil {
		return fmt.Errorf("error saving destination bundle archive: %w", err)
	}

	return nil
}

type ImageArchiveOpener struct {
	Archive  *BundleArchive
	ImageTag string
}

func NewImageArchiveOpener(archive *BundleArchive, imageTag string) *ImageArchiveOpener {
	return &ImageArchiveOpener{
		Archive:  archive,
		ImageTag: imageTag,
	}
}

func (opener *ImageArchiveOpener) Open() (io.ReadCloser, error) {
	return opener.Archive.Reader.ReadImageArchive(opener.ImageTag)
}

type ImageArchiveReadCloser struct {
	reader io.Reader
	closer func() error
}

func NewImageArchiveReadCloser(reader io.Reader, closer func() error) *ImageArchiveReadCloser {
	return &ImageArchiveReadCloser{
		reader: reader,
		closer: closer,
	}
}

func (opener *ImageArchiveReadCloser) Read(p []byte) (int, error) {
	return opener.reader.Read(p)
}

func (opener *ImageArchiveReadCloser) Close() error {
	return opener.closer()
}
