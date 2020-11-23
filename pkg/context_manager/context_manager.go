package context_manager

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/pkg/werf"
)

func GetContextTmpDir() string {
	return filepath.Join(werf.GetServiceDir(), "tmp", "context")
}

func ContextAddFileChecksum(ctx context.Context, contextAddFile []string, projectDir string) (string, error) {
	logboek.Context(ctx).Debug().LogF("-- ContextAddFileChecksum %q %q\n", projectDir, contextAddFile)

	h := sha256.New()

	for _, addFile := range contextAddFile {
		h.Write([]byte(addFile))

		path := filepath.Join(projectDir, addFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return "", fmt.Errorf("error accessing %q: %s", path, err)
		}

		if f, err := os.Open(path); err != nil {
			return "", fmt.Errorf("unable to open %q: %s", path, err)
		} else {
			defer f.Close()
			if _, err := io.Copy(h, f); err != nil {
				return "", fmt.Errorf("error reading %q: %s", path, err)
			}
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type contextAddFileDescriptor struct {
	AddFile           string
	PathInsideContext string
}

func ApplyContextAddFileToArchive(ctx context.Context, originalArchivePath string, contextPath string, contextAddFile []string, projectDir string) (string, error) {
	path := filepath.Join(GetContextTmpDir(), uuid.NewV4().String())
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return "", fmt.Errorf("unable to create dir %q: %s", filepath.Dir(path), err)
	}

	logboek.Context(ctx).Default().LogF("Will copy %q archive to %q\n", originalArchivePath, path)

	source, err := os.Open(originalArchivePath)
	if err != nil {
		return "", fmt.Errorf("unable to open %q: %s", originalArchivePath, err)
	}
	defer source.Close()

	destination, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("unable to create %q: %s", path, err)
	}
	defer destination.Close()

	tr := tar.NewReader(source)
	tw := tar.NewWriter(destination)
	defer tw.Close()

	var contextAddFileDescriptors []*contextAddFileDescriptor
	for _, addFile := range contextAddFile {
		var destFilePath string
		if contextPath != "" {
			if !util.IsSubpathOfBasePath(contextPath, addFile) {
				return "", fmt.Errorf("specified contextAddFile %q is out of context %q", addFile, contextPath)
			}
			destFilePath = util.GetRelativeToBaseFilepath(contextPath, addFile)
		} else {
			destFilePath = addFile
		}

		contextAddFileDescriptors = append(contextAddFileDescriptors, &contextAddFileDescriptor{
			AddFile:           addFile,
			PathInsideContext: destFilePath,
		})
	}

CopyArchive:
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", fmt.Errorf("error reading archive %q: %s", originalArchivePath, err)
		}

		for _, addFileDesc := range contextAddFileDescriptors {
			if hdr.Name == filepath.ToSlash(addFileDesc.PathInsideContext) {
				logboek.Context(ctx).Default().LogF("Matched file %q for replacement in the archive %q by contextAddFile=%q directive\n", hdr.Name, path, addFileDesc.AddFile)
				continue CopyArchive
			}
		}

		tw.WriteHeader(hdr)

		if _, err := io.Copy(tw, tr); err != nil {
			return "", fmt.Errorf("error copying %q from %q archive to %q: %s", hdr.Name, originalArchivePath, path, err)
		}

		logboek.Context(ctx).Default().LogF("Copied %s from %q archive to %q\n", hdr.Name, originalArchivePath, path)
	}

	for _, addFileDesc := range contextAddFileDescriptors {
		sourceFilePath := filepath.Join(projectDir, addFileDesc.AddFile)
		tarEntryName := filepath.ToSlash(addFileDesc.PathInsideContext)
		if err := copyFileIntoTar(sourceFilePath, tarEntryName, tw); err != nil {
			return "", fmt.Errorf("unable to copy %q from workinto archive %q: %s", sourceFilePath, path, err)
		}
		logboek.Context(ctx).Default().LogF("Copied file %q in the archive %q with %q file from working directory (contextAddFile=%s directive)\n", tarEntryName, path, sourceFilePath, addFileDesc.AddFile)
	}

	return path, nil
}

func copyFileIntoTar(sourceFilePath string, tarEntryName string, tw *tar.Writer) error {
	sourceFileStat, err := os.Lstat(sourceFilePath)
	if err != nil {
		return fmt.Errorf("error accessing %q stat: %s", sourceFilePath, err)
	}

	isSymlink := sourceFileStat.Mode()&os.ModeSymlink != 0
	if isSymlink {
		linkname, err := os.Readlink(sourceFilePath)
		if err != nil {
			return fmt.Errorf("cannot read symlink %q: %s", sourceFilePath, err)
		}

		if err := tw.WriteHeader(&tar.Header{
			Format:     tar.FormatGNU,
			Typeflag:   tar.TypeSymlink,
			Name:       tarEntryName,
			Linkname:   linkname,
			Mode:       int64(sourceFileStat.Mode()),
			Size:       sourceFileStat.Size(),
			ModTime:    sourceFileStat.ModTime(),
			AccessTime: sourceFileStat.ModTime(),
			ChangeTime: sourceFileStat.ModTime(),
		}); err != nil {
			return fmt.Errorf("unable to write tar symlink header for file %s: %s", tarEntryName, err)
		}
	} else if sourceFileStat.Mode().IsRegular() {
		if err := tw.WriteHeader(&tar.Header{
			Format:     tar.FormatGNU,
			Name:       tarEntryName,
			Mode:       int64(sourceFileStat.Mode()),
			Size:       sourceFileStat.Size(),
			ModTime:    sourceFileStat.ModTime(),
			AccessTime: sourceFileStat.ModTime(),
			ChangeTime: sourceFileStat.ModTime(),
		}); err != nil {
			return fmt.Errorf("unable to write tar header for file %q: %s", tarEntryName, err)
		}

		f, err := os.Open(sourceFilePath)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %s", sourceFilePath, err)
		}

		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("unable to write data to tar archive from file %q: %s", sourceFilePath, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing file %q: %s", sourceFilePath, err)
		}
	}

	return nil
}
