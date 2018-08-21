package git_repo

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage"
)

type Archive struct {
	PathFilter PathFilter
	Repo       struct {
		Tree   *object.Tree
		Storer storage.Storer
	}
}

func (a *Archive) Type() (ArchiveType, error) {
	treeWalker := object.NewTreeWalker(a.Repo.Tree, true, nil)

	basePath := NormalizeAbsolutePath(a.PathFilter.BasePath)

	if basePath == "/" {
		return DirectoryArchive, nil
	}

	for {
		name, entry, err := treeWalker.Next()
		if err == io.EOF {
			break
		}

		if NormalizeAbsolutePath(name) == basePath {
			if entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule {
				return DirectoryArchive, nil
			}
			return FileArchive, nil
		}
	}

	return "", fmt.Errorf("cannot find base path `%s` entry in repo", a.PathFilter.BasePath)
}

func startMemprofile() {
	runtime.MemProfileRate = 1
}

func stopMemprofile() {
	memprofilePath := fmt.Sprintf("/tmp/create-tar-memprofile-%s", uuid.NewV4())
	fmt.Printf("Creating mem profile: %s\n", memprofilePath)
	f, err := os.Create(memprofilePath)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	f.Close()
}

func (a *Archive) CreateTar(output io.Writer) error {
	// startMemprofile()
	// defer stopMemprofile()

	tw := tar.NewWriter(output)
	treeWalker := object.NewTreeWalker(a.Repo.Tree, true, nil)

	var err error

	err = a.writeEntriesToArchive(tw, treeWalker)
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	return nil
}

func (a *Archive) writeEntriesToArchive(tw *tar.Writer, treeWalker *object.TreeWalker) error {
	now := time.Now()
	chunkBuf := make([]byte, 16*1024*1024) // 16Mb chunk

	for {
		name, entry, err := treeWalker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule {
			continue
		}
		if !a.PathFilter.IsFilePathValid(name) {
			continue
		}

		// NOTICE: Current GetBlob implementation indirectly reading file content.
		// NOTICE: Which cause big memory usage on big repos.
		// NOTICE: Also this is a execution speed bottleneck for big repos.
		// NOTICE: See go-git issue https://github.com/src-d/go-git/issues/832.
		blob, err := object.GetBlob(a.Repo.Storer, entry.Hash)
		if err != nil {
			return err
		}
		blobReader, err := blob.Reader()
		if err != nil {
			return err
		}

		filename := a.PathFilter.TrimFileBasePath(name)

		if entry.Mode == filemode.Symlink {
			buf := bytes.Buffer{}
			_, readErr := buf.ReadFrom(blobReader)
			if readErr != nil {
				return readErr
			}
			linkname := buf.String()

			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       filename,
				Mode:       int64(filemode.Symlink),
				Linkname:   linkname,
				Size:       blob.Size,
				ModTime:    now,
				AccessTime: now,
				ChangeTime: now,
			})
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header: %s", err)
			}
		} else {
			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Name:       filename,
				Mode:       int64(entry.Mode),
				Size:       blob.Size,
				ModTime:    now,
				AccessTime: now,
				ChangeTime: now,
			})
			if err != nil {
				return fmt.Errorf("unable to write tar header: %s", err)
			}

			err = ReadChunks(chunkBuf, blobReader, func(bytes []byte) error {
				_, writeErr := tw.Write(bytes)
				if writeErr != nil {
					return fmt.Errorf("unable to write data to tar archive: %s", writeErr)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Archive) IsAnyEntries() (bool, error) {
	treeWalker := object.NewTreeWalker(a.Repo.Tree, true, nil)

	for {
		name, entry, err := treeWalker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}

		if entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule {
			continue
		}

		if !a.PathFilter.IsFilePathValid(name) {
			continue
		}

		return true, nil
	}

	return false, nil
}

func ReadChunks(chunkBuf []byte, reader io.Reader, handleChunk func(bytes []byte) error) error {
	for {
		n, err := reader.Read(chunkBuf)

		if n > 0 {
			handleErr := handleChunk(chunkBuf[:n])
			if handleErr != nil {
				return handleErr
			}
		}

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
	}
}
