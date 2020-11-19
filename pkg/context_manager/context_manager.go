package context_manager

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

func LsTree(gitLsTreeResult, contextAddFile []string) error {
	return nil
}

func ApplyContextAddFileToArchive(archivePath string, context string, contextAddFile []string, projectDir string) error {
	fmt.Printf("-- ApplyContextAddFileToArchive %q %q %v %q\n", archivePath, context, contextAddFile, projectDir)

	file, err := os.OpenFile(archivePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to open %q: %s", archivePath, err)
	}
	if _, err := file.Seek(-2<<9, os.SEEK_END); err != nil {
		return err
	}

	tw := tar.NewWriter(file)

	for _, addFile := range contextAddFile {
		// TODO: raise an error when specified addfile is out of context
		// TODO: check addfile is not a directory, raise an error in this case

		sourceFilePath := filepath.Join(projectDir, addFile)

		var destFilePath string
		if context != "" {
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
		} else {
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
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("cannot write tar archive: %s", err)
	}

	return nil
}
