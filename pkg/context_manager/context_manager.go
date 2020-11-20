package context_manager

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/util"
)

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

func ApplyContextAddFileToArchive(ctx context.Context, archivePath string, context string, contextAddFile []string, projectDir string) error {
	logboek.Context(ctx).Debug().LogF("-- ApplyContextAddFileToArchive %q %q %v %q\n", archivePath, context, contextAddFile, projectDir)

	file, err := os.OpenFile(archivePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to open %q: %s", archivePath, err)
	}
	if _, err := file.Seek(-2<<9, os.SEEK_END); err != nil {
		return err
	}

	tw := tar.NewWriter(file)

	for _, addFile := range contextAddFile {
		sourceFilePath := filepath.Join(projectDir, addFile)

		var destFilePath string
		if context != "" {
			if !util.IsSubpathOfBasePath(context, addFile) {
				return fmt.Errorf("specified contextAddFile %q is out of context %q", addFile, context)
			}
			destFilePath = util.GetRelativeToBaseFilepath(context, addFile)
		} else {
			destFilePath = addFile
		}
		tarEntryName := filepath.ToSlash(destFilePath)

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
		} else {
			return fmt.Errorf("unexpected contextAddFile %q file type %x: only regular files or symlinks are supported", addFile, sourceFileStat.Mode())
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("cannot write tar archive: %s", err)
	}

	return nil
}
