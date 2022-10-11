package util

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"

	"github.com/werf/logboek"
)

type CreateArchiveOptions struct {
	CopyTarOptions
	AfterCopyFunc func(tw *tar.Writer) error
}

func CreateArchiveBasedOnAnotherOne(ctx context.Context, sourceArchivePath, destinationArchivePath string, opts CreateArchiveOptions) error {
	return CreateArchive(destinationArchivePath, func(tw *tar.Writer) error {
		source, err := os.Open(sourceArchivePath)
		if err != nil {
			return fmt.Errorf("unable to open %q: %w", sourceArchivePath, err)
		}
		defer source.Close()

		if err := CopyTar(ctx, source, tw, opts.CopyTarOptions); err != nil {
			return err
		}

		if opts.AfterCopyFunc != nil {
			return opts.AfterCopyFunc(tw)
		}

		return nil
	})
}

func CreateArchive(archivePath string, f func(tw *tar.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o777); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", filepath.Dir(archivePath), err)
	}

	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("unable to create %q: %w", archivePath, err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	return f(tw)
}

func CopyFileIntoTar(tw *tar.Writer, tarEntryName, filePath string) error {
	stat, err := os.Lstat(filePath)
	if err != nil {
		return fmt.Errorf("unable to stat file %q: %w", filePath, err)
	}

	if stat.Mode().IsDir() {
		return fmt.Errorf("directory %s cannot be added to tar archive", filePath)
	}

	header := &tar.Header{
		Name:       tarEntryName,
		Mode:       int64(stat.Mode()),
		Size:       stat.Size(),
		ModTime:    stat.ModTime(),
		AccessTime: stat.ModTime(),
		ChangeTime: stat.ModTime(),
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		linkname, err := os.Readlink(filePath)
		if err != nil {
			return fmt.Errorf("unable to read link %q: %w", filePath, err)
		}

		header.Linkname = linkname
		header.Typeflag = tar.TypeSymlink
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write tar header for file %s: %w", tarEntryName, err)
	}

	if stat.Mode().IsRegular() {
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", filePath, err)
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil
		}

		if _, err := tw.Write(data); err != nil {
			return fmt.Errorf("unable to write data to tar archive from file %q: %w", filePath, err)
		}
	}

	return nil
}

func CopyGitIndexEntryIntoTar(tw *tar.Writer, tarEntryName string, entry *index.Entry, obj plumbing.EncodedObject) error {
	r, err := obj.Reader()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:       tarEntryName,
		Mode:       int64(entry.Mode),
		Size:       int64(entry.Size),
		ModTime:    entry.ModifiedAt,
		AccessTime: entry.ModifiedAt,
		ChangeTime: entry.ModifiedAt,
	}

	if entry.Mode == filemode.Symlink {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		linkname := string(data)
		header.Linkname = linkname
		header.Typeflag = tar.TypeSymlink
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("unable to write tar header for git index entry %s: %w", tarEntryName, err)
	}

	if entry.Mode.IsFile() {
		if _, err := io.Copy(tw, r); err != nil {
			return fmt.Errorf("unable to write data to tar for git index entry %q: %w", tarEntryName, err)
		}
	}

	return nil
}

type CopyTarOptions struct {
	IncludePaths []string
	ExcludePaths []string
}

func CopyTar(ctx context.Context, in io.Reader, tw *tar.Writer, opts CopyTarOptions) error {
	tr := tar.NewReader(in)

ArchiveCopying:
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("unable to read archive: %w", err)
		}

		for _, excPath := range opts.ExcludePaths {
			if hdr.Name == filepath.ToSlash(excPath) {
				if debugArchiveUtil() {
					logboek.Context(ctx).Debug().LogF("Source archive file excluded: %q\n", hdr.Name)
				}

				continue ArchiveCopying
			}
		}

		if len(opts.IncludePaths) > 0 {
			for _, incPath := range opts.IncludePaths {
				if hdr.Name == filepath.ToSlash(incPath) {
					if debugArchiveUtil() {
						logboek.Context(ctx).Debug().LogF("Source archive file included: %q\n", hdr.Name)
					}
					goto CopyEntry
				}
			}

			continue ArchiveCopying
		}

	CopyEntry:
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("unable to write tar header entry %q: %w", hdr.Name, err)
		}

		if _, err := io.Copy(tw, tr); err != nil {
			return fmt.Errorf("unable to copy tar entry %q data: %w", hdr.Name, err)
		}

		if debugArchiveUtil() {
			logboek.Context(ctx).Debug().LogF("Source archive file was added: %q\n", hdr.Name)
		}
	}

	return nil
}

type ExtractTarOptions struct {
	UID, GID *uint32
}

func ExtractTar(tarFileReader io.Reader, dstDir string, opts ExtractTarOptions) error {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create dir %q: %w", dstDir, err)
	}
	if err := Chown(dstDir, opts.UID, opts.GID); err != nil {
		return fmt.Errorf("unable to chown dir %q: %w", dstDir, err)
	}

	tarReader := tar.NewReader(tarFileReader)
	for {
		tarEntryHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to Next() while extracting tar: %w", err)
		}

		tarEntryPath := filepath.Join(dstDir, tarEntryHeader.Name)
		tarEntryFileInfo := tarEntryHeader.FileInfo()

		switch tarEntryHeader.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(tarEntryPath, tarEntryFileInfo.Mode()); err != nil {
				return fmt.Errorf("unable to create new dir %q while extracting tar: %w", tarEntryPath, err)
			}
		case tar.TypeBlock, tar.TypeChar, tar.TypeReg, tar.TypeFifo:
			if err := os.MkdirAll(filepath.Dir(tarEntryPath), os.ModePerm); err != nil {
				return fmt.Errorf("unable to create new directory %q while extracting tar: %w", filepath.Dir(tarEntryPath), err)
			}

			file, err := os.OpenFile(tarEntryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, tarEntryFileInfo.Mode())
			if err != nil {
				return fmt.Errorf("unable to create new file %q while extracting tar: %w", tarEntryPath, err)
			}
			defer file.Close()

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return fmt.Errorf("unable to create file %q while extracting tar: %w", tarEntryPath, err)
			}
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(tarEntryPath), os.ModePerm); err != nil {
				return fmt.Errorf("unable to create new directory %q while extracting tar: %w", filepath.Dir(tarEntryPath), err)
			}

			if err := os.Link(tarEntryHeader.Linkname, tarEntryPath); err != nil {
				return fmt.Errorf("unable to create hard link %q while extracting tar: %w", tarEntryPath, err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(tarEntryPath), os.ModePerm); err != nil {
				return fmt.Errorf("unable to create new directory %q while extracting tar: %w", filepath.Dir(tarEntryPath), err)
			}

			if err := os.Symlink(tarEntryHeader.Linkname, tarEntryPath); err != nil {
				return fmt.Errorf("unable to create symlink %q while extracting tar: %w", tarEntryPath, err)
			}
		default:
			return fmt.Errorf("tar entry %q of unexpected type: %b", tarEntryHeader.Name, tarEntryHeader.Typeflag)
		}

		for _, p := range FilepathsWithParents(tarEntryHeader.Name) {
			if err := Chown(filepath.Join(dstDir, p), opts.UID, opts.GID); err != nil {
				return fmt.Errorf("unable to chown file %q: %w", p, err)
			}
		}
	}

	return nil
}

func Chown(path string, uid, gid *uint32) error {
	if uid != nil || gid != nil {
		osUid := -1
		osGid := -1

		if uid != nil {
			osUid = int(*uid)
		}
		if gid != nil {
			osGid = int(*gid)
		}

		if err := os.Chown(path, osUid, osGid); err != nil {
			return fmt.Errorf("unable to set owner and group to %d:%d for file %q: %w", osUid, osGid, path, err)
		}
	}

	return nil
}

func WriteDirAsTar(dir string, w io.Writer) error {
	tarWriter := tar.NewWriter(w)

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %q: %w", path, err)
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if debugArchiveUtil() {
			fmt.Printf("filepath.Walk %q\n", relPath)
		}

		if info.Mode().IsDir() {
			header := &tar.Header{
				Name:     relPath,
				Size:     info.Size(),
				Mode:     int64(info.Mode()),
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeDir,
			}

			err = tarWriter.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("could not tar write header for %q: %w", path, err)
			}

			if debugArchiveUtil() {
				fmt.Printf("Written dir %q\n", relPath)
			}

			return nil
		}

		header := &tar.Header{
			Name:     relPath,
			Size:     info.Size(),
			Mode:     int64(info.Mode()),
			ModTime:  info.ModTime(),
			Typeflag: tar.TypeReg,
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("could not tar write header for %q: %w", path, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open %q: %w", path, err)
		}

		n, err := io.Copy(tarWriter, file)
		if err != nil {
			return fmt.Errorf("unable to write %q into tar: %w", path, err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close %q: %w", path, err)
		}

		if debugArchiveUtil() {
			fmt.Printf("Written file %q (%d bytes)\n", relPath, n)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("unable to close tar writer: %w", err)
	}

	return nil
}

func debugArchiveUtil() bool {
	return os.Getenv("WERF_DEBUG_ARCHIVE_UTIL") == "1"
}
