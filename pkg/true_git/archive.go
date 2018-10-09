package true_git

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type ArchiveOptions struct {
	Commit     string
	PathFilter PathFilter
}

type ArchiveDescriptor struct {
	Type    ArchiveType
	IsEmpty bool
}

type ArchiveType string

const (
	FileArchive      ArchiveType = "file"
	DirectoryArchive ArchiveType = "directory"
)

func ArchiveWithSubmodules(out io.Writer, gitDir, workTreeDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	return writeArchive(out, gitDir, workTreeDir, true, opts)
}

func Archive(out io.Writer, gitDir, workTreeDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	return writeArchive(out, gitDir, workTreeDir, false, opts)
}

func debugArchive() bool {
	return os.Getenv("DAPP_TRUE_GIT_DEBUG_ARCHIVE") == "1"
}

func writeArchive(out io.Writer, gitDir, workTreeDir string, withSubmodules bool, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir `%s`: %s", gitDir, err)
	}

	workTreeDir, err = filepath.Abs(workTreeDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree dir `%s`: %s", workTreeDir, err)
	}

	if withSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return nil, err
		}
	}

	err = switchWorkTree(gitDir, workTreeDir, opts.Commit)
	if err != nil {
		return nil, fmt.Errorf("cannot reset work tree `%s` to commit `%s`: %s", workTreeDir, opts.Commit, err)
	}

	if withSubmodules {
		err := updateSubmodules(gitDir, workTreeDir)
		if err != nil {
			return nil, fmt.Errorf("cannot update submodules: %s", err)
		}
	}

	desc := &ArchiveDescriptor{
		IsEmpty: true,
	}

	tw := tar.NewWriter(out)

	err = filepath.Walk(workTreeDir, func(absPath string, info os.FileInfo, accessErr error) error {
		if accessErr != nil {
			return fmt.Errorf("error accessing `%s`: %s", absPath, accessErr)
		}

		baseName := filepath.Base(absPath)
		for _, p := range []string{".git", ".gitmodules", ".gitkeep", ".gitignore"} {
			if baseName == p {
				return nil
			}
		}

		var path string
		if absPath == workTreeDir {
			path = "."
		} else {
			path = strings.TrimPrefix(absPath, NormalizeDirectoryPrefix(workTreeDir))
		}

		if NormalizeAbsolutePath(path) == NormalizeAbsolutePath(opts.PathFilter.BasePath) {
			if info.IsDir() {
				desc.Type = DirectoryArchive

				if debugArchive() {
					fmt.Printf("Found BasePath `%s` directory: directory archive type\n", path)
				}
			} else {
				desc.Type = FileArchive

				if debugArchive() {
					fmt.Printf("Found BasePath `%s` file: file archive\n", path)
				}
			}
		}

		if info.IsDir() {
			return nil
		}

		if !opts.PathFilter.IsFilePathValid(path) {
			if debugArchive() {
				fmt.Printf("Excluded path `%s` by path filter %s\n", path, opts.PathFilter.String())
			}
			return nil
		}

		archivePath := opts.PathFilter.TrimFileBasePath(path)

		desc.IsEmpty = false

		if info.Mode()&os.ModeSymlink != 0 {
			linkname, err := os.Readlink(absPath)
			if err != nil {
				return fmt.Errorf("cannot read symlink `%s`: %s", absPath, err)
			}

			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       archivePath,
				Linkname:   string(linkname),
				Mode:       int64(info.Mode()),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header for file `%s`: %s", archivePath, err)
			}

			if debugArchive() {
				fmt.Printf("Added archive symlink `%s` -> `%s`\n", path, linkname)
			}

			return nil
		}

		err = tw.WriteHeader(&tar.Header{
			Format:     tar.FormatGNU,
			Name:       archivePath,
			Mode:       int64(info.Mode()),
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			AccessTime: info.ModTime(),
			ChangeTime: info.ModTime(),
		})
		if err != nil {
			return fmt.Errorf("unable to write tar header for file `%s`: %s", archivePath, err)
		}

		file, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("unable to open file `%s`: %s", absPath, err)
		}

		_, err = io.Copy(tw, file)
		if err != nil {
			return fmt.Errorf("unable to write data to tar archive from file `%s`: %s", path, err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("error closing file `%s`: %s", absPath, err)
		}

		if debugArchive() {
			fmt.Printf("Added archive file `%s`\n", path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("entries iteration failed in `%s`: %s", workTreeDir, err)
	}

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot write tar archive: %s", err)
	}

	if desc.Type == "" {
		return nil, fmt.Errorf("base path `%s` entry not found repo", opts.PathFilter.BasePath)
	}

	return desc, nil
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
