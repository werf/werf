package git

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type ArchiveOptions struct {
	Commit         string
	PathFilter     PathFilter
	WithSubmodules bool
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

func Archive(out io.Writer, repoPath string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	clonePath := filepath.Join("/tmp", fmt.Sprintf("git-clone-%s", uuid.NewV4().String()))
	defer os.RemoveAll(clonePath)

	cmd := exec.Command("git", "clone", repoPath, clonePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("temporary clone creation failed: %s\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", clonePath, "checkout", opts.Commit)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot checkout commit `%s`: %s\n%s", opts.Commit, err, output)
	}

	if opts.WithSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return nil, err
		}

		cmd = exec.Command("git", "-C", clonePath, "submodule", "update", "--init", "--recursive")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("cannot update submodules")
		}
	}

	desc := &ArchiveDescriptor{
		IsEmpty: true,
	}

	now := time.Now()
	tw := tar.NewWriter(out)

	trees := []struct{ Path, Commit string }{
		{"/", opts.Commit},
	}

	for len(trees) > 0 {
		tree := trees[0]
		trees = trees[1:]

		cmd = exec.Command(
			"git", "-C", filepath.Join(clonePath, tree.Path),
			"ls-tree", "--long", "--full-tree", "-r", "-z", "-t",
			tree.Commit,
		)

		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("cannot list git tree: %s\n%s", err, output)
		}

		lines := strings.Split(string(output), "\000")
		lines = lines[:len(lines)-1]

	HandleEntries:
		for _, line := range lines {
			fields := strings.SplitN(line, " ", 4)

			rawMode := strings.TrimLeft(fields[0], " ")
			objectType := strings.TrimLeft(fields[1], " ")
			objectID := strings.TrimLeft(fields[2], " ")

			fields = strings.SplitN(strings.TrimLeft(fields[3], " "), "\t", 2)

			rawObjectSize := fields[0]
			filePath := filepath.Join(tree.Path, fields[1])
			fullFilePath := filepath.Join(clonePath, filePath)

			mode, err := strconv.ParseInt(rawMode, 8, 64)
			if err != nil {
				return nil, fmt.Errorf("unexpected git ls-tree file mode `%s`: %s", rawMode, err)
			}

			switch objectType {
			case "blob":
				if !opts.PathFilter.IsFilePathValid(filePath) {
					continue HandleEntries
				}
				archiveFilePath := opts.PathFilter.TrimFileBasePath(filePath)

				if NormalizeAbsolutePath(filePath) == NormalizeAbsolutePath(opts.PathFilter.BasePath) {
					desc.Type = FileArchive
				}

				size, err := strconv.ParseInt(rawObjectSize, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("unexpected git ls-tree file size `%s`: %s", rawObjectSize, err)
				}

				if mode == 0120000 { // symlink
					linkname, err := os.Readlink(fullFilePath)
					if err != nil {
						return nil, fmt.Errorf("cannot read symlink `%s`: %s", fullFilePath, err)
					}

					err = tw.WriteHeader(&tar.Header{
						Format:     tar.FormatGNU,
						Typeflag:   tar.TypeSymlink,
						Name:       archiveFilePath,
						Mode:       mode,
						Linkname:   string(linkname),
						Size:       size,
						ModTime:    now,
						AccessTime: now,
						ChangeTime: now,
					})
					if err != nil {
						return nil, fmt.Errorf("unable to write tar symlink header: %s", err)
					}
				} else {
					err = tw.WriteHeader(&tar.Header{
						Format:     tar.FormatGNU,
						Name:       archiveFilePath,
						Mode:       mode,
						Size:       size,
						ModTime:    now,
						AccessTime: now,
						ChangeTime: now,
					})
					if err != nil {
						return nil, fmt.Errorf("unable to write tar header: %s", err)
					}

					file, err := os.Open(fullFilePath)
					if err != nil {
						return nil, fmt.Errorf("unable to open sss file `%s`: %s", fullFilePath, err)
					}

					_, err = io.Copy(tw, file)
					if err != nil {
						return nil, fmt.Errorf("unable to write data to tar archive: %s", err)
					}
				}

			case "commit":
				if opts.WithSubmodules {
					trees = append(trees, struct{ Path, Commit string }{filePath, objectID})
				}

				if NormalizeAbsolutePath(filePath) == NormalizeAbsolutePath(opts.PathFilter.BasePath) {
					desc.Type = DirectoryArchive
				}

			case "tree":
				if NormalizeAbsolutePath(filePath) == NormalizeAbsolutePath(opts.PathFilter.BasePath) {
					desc.Type = DirectoryArchive
				}

			default:
				panic(fmt.Sprintf("unexpected object type `%s`", objectType))
			}

			desc.IsEmpty = false
		}
	}

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot write tar archive: %s", err)
	}

	if NormalizeAbsolutePath(opts.PathFilter.BasePath) == "/" {
		desc.Type = DirectoryArchive
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
