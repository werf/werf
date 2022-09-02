package bundles

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

const (
	chartArchiveFileName = "chart.tar.gz"
)

type BundleWriter interface {
	Open() error
	WriteChartArchive(data []byte) error
	WriteImageArchive(imageTag string, data []byte) error
	Save() error
}

type BundleReader interface {
	ReadChartArchive() ([]byte, error)
	GetImageArchiveOpener(imageTag string) *ImageArchiveOpener
	ReadImageArchive(imageTag string) (*ImageArchiveReadCloser, error)
}

type BundleArchive struct {
	Path string

	tmpArchivePath   string
	tmpArchiveWriter *tar.Writer
	tmpArchiveCloser func() error
}

func NewBundleArchive(path string) *BundleArchive {
	return &BundleArchive{
		Path: path,
	}
}

func (bundle *BundleArchive) Open() error {
	p := fmt.Sprintf("%s.%s.tmp", bundle.Path, uuid.New().String())

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("unable to open tmp archive file %q: %w", p, err)
	}

	zipper := gzip.NewWriter(f)
	zipper.Header.Comment = "bundle-archive"
	twriter := tar.NewWriter(zipper)

	bundle.tmpArchivePath = p
	bundle.tmpArchiveWriter = twriter
	bundle.tmpArchiveCloser = func() error {
		if err := twriter.Close(); err != nil {
			return fmt.Errorf("unable to close tar writer for %q: %w", bundle.tmpArchivePath, err)
		}
		if err := zipper.Close(); err != nil {
			return fmt.Errorf("unable to close zipper for %q: %w", bundle.tmpArchivePath, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", bundle.tmpArchivePath, err)
		}
		return nil
	}

	now := time.Now()
	header := &tar.Header{
		Name:       "images",
		Typeflag:   tar.TypeDir,
		Mode:       0o777,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
	if err := bundle.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write images dir header: %w", err)
	}

	return nil
}

func (bundle *BundleArchive) Save() error {
	if bundle.tmpArchiveWriter == nil {
		panic(fmt.Sprintf("bundle archive %q is not opened", bundle.Path))
	}

	if err := bundle.tmpArchiveCloser(); err != nil {
		return fmt.Errorf("unable to close tmp archive %q: %w", bundle.tmpArchivePath, err)
	}

	if err := os.RemoveAll(bundle.Path); err != nil {
		return fmt.Errorf("unable to cleanup destination archive path %q: %w", bundle.Path, err)
	}

	if err := os.Rename(bundle.tmpArchivePath, bundle.Path); err != nil {
		return fmt.Errorf("unable to rename tmp bundle archive %q to %q: %w", bundle.tmpArchivePath, bundle.Path, err)
	}

	return nil
}

func (bundle *BundleArchive) WriteChartArchive(data []byte) error {
	now := time.Now()
	header := &tar.Header{
		Name:       chartArchiveFileName,
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       int64(len(data)),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := bundle.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write %q header: %w", chartArchiveFileName, err)
	}

	if _, err := bundle.tmpArchiveWriter.Write(data); err != nil {
		return fmt.Errorf("unable to write %q data: %w", chartArchiveFileName, err)
	}

	return nil
}

func (bundle *BundleArchive) WriteImageArchive(imageTag string, data []byte) error {
	now := time.Now()

	header := &tar.Header{
		Name:       fmt.Sprintf("images/%s.tar.gz", imageTag),
		Typeflag:   tar.TypeReg,
		Mode:       0o777,
		Size:       int64(len(data)),
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}

	if err := bundle.tmpArchiveWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write chart.tar.gz header: %w", err)
	}

	if _, err := bundle.tmpArchiveWriter.Write(data); err != nil {
		return fmt.Errorf("unable to write chart.tar.gz data: %w", err)
	}

	return nil
}

func (bundle *BundleArchive) openForReading() (*tar.Reader, func() error, error) {
	f, err := os.Open(bundle.Path)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	unzipper, err := gzip.NewReader(f)
	if err != nil {
		return nil, f.Close, fmt.Errorf("unable to open bundle archive gzip %q: %w", bundle.Path, err)
	}

	closer := func() error {
		if err := unzipper.Close(); err != nil {
			return fmt.Errorf("unable to close gzipper for %q: %w", bundle.Path, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", bundle.Path, err)
		}
		return nil
	}

	return tar.NewReader(unzipper), closer, nil
}

func (bundle *BundleArchive) ReadChartArchive() ([]byte, error) {
	treader, closer, err := bundle.openForReading()
	defer closer()

	if err != nil {
		return nil, fmt.Errorf("unable to open bundle archive: %w", err)
	}

	b := bytes.NewBuffer(nil)

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no chart archive found in the bundle archive %q", bundle.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}
		if header.Name != chartArchiveFileName {
			continue
		}

		if _, err := io.Copy(b, treader); err != nil {
			return nil, fmt.Errorf("unable to read chart archive %q from the bundle archive %q: %w", chartArchiveFileName, bundle.Path, err)
		}

		return b.Bytes(), nil
	}
}

func (bundle *BundleArchive) GetImageArchiveOpener(imageTag string) *ImageArchiveOpener {
	return NewImageArchiveOpener(bundle, imageTag)
}

func (bundle *BundleArchive) ReadImageArchive(imageTag string) (*ImageArchiveReadCloser, error) {
	treader, closer, err := bundle.openForReading()
	if err != nil {
		defer closer()
		return nil, fmt.Errorf("unable to open bundle archive: %w", err)
	}

	for {
		header, err := treader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("no image tag %q found in the bundle archive %q", imageTag, bundle.Path)
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if header.Name == fmt.Sprintf("images/%s.tar.gz", imageTag) {
			unzipper, err := gzip.NewReader(treader)
			if err != nil {
				return nil, fmt.Errorf("unable to create gzip reader for image archive: %w", err)
			}

			return NewImageArchiveReadCloser(unzipper, func() error {
				if err := unzipper.Close(); err != nil {
					return fmt.Errorf("unable to close gzip reader for image archive: %w", err)
				}
				return closer()
			}), nil
		}
	}
}

func (bundle *BundleArchive) ReadChart(ctx context.Context) (*chart.Chart, error) {
	chartBytes, err := bundle.ReadChartArchive()
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

	if err := bundle.WriteChartArchive(chartBytes); err != nil {
		return fmt.Errorf("unable to write chart archive into bundle archive: %w", err)
	}

	return nil
}

func (bundle *BundleArchive) CopyTo(ctx context.Context, to BundleAccessor) error {
	return to.CopyFromArchive(ctx, bundle)
}

func (bundle *BundleArchive) CopyFromArchive(ctx context.Context, fromArchive *BundleArchive) error {
	if err := copy.Copy(fromArchive.Path, bundle.Path); err != nil {
		return fmt.Errorf("unable to copy file %q to %q: %w", fromArchive.Path, bundle.Path, err)
	}
	return nil
}

func (bundle *BundleArchive) CopyFromRemote(ctx context.Context, fromRemote *RemoteBundle) error {
	ch, err := fromRemote.ReadChart(ctx)
	if err != nil {
		return fmt.Errorf("unable to read chart from remote bundle: %w", err)
	}

	if err := bundle.Open(); err != nil {
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
					zipper := gzip.NewWriter(imageBytes)

					if err := fromRemote.RegistryClient.PullImageArchive(ctx, zipper, imageRef); err != nil {
						return fmt.Errorf("error pulling image %q archive: %w", imageRef, err)
					}

					if err := zipper.Close(); err != nil {
						return fmt.Errorf("unable to close gzip writer: %w", err)
					}

					if err := bundle.WriteImageArchive(tag, imageBytes.Bytes()); err != nil {
						return fmt.Errorf("error writing image %q into bundle archive: %w", imageRef, err)
					}
				} else {
					return fmt.Errorf("unexpected value .Values.werf.image.%s=%v", imageName, v)
				}
			}
		}
	}

	if err := bundle.Save(); err != nil {
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
	return opener.Archive.ReadImageArchive(opener.ImageTag)
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
