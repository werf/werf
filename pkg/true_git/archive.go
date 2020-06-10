package true_git

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/logboek"

	"github.com/go-git/go-git/v5/plumbing/filemode"

	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/true_git/ls_tree"
	"github.com/werf/werf/pkg/util"
)

type ArchiveOptions struct {
	Commit      string
	PathMatcher path_matcher.PathMatcher
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

func ArchiveWithSubmodules(out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var res *ArchiveDescriptor

	err := withWorkTreeCacheLock(workTreeCacheDir, func() error {
		writeArchiveRes, err := writeArchive(out, gitDir, workTreeCacheDir, true, opts)
		res = writeArchiveRes
		return err
	})

	return res, err
}

func Archive(out io.Writer, gitDir, workTreeCacheDir string, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var res *ArchiveDescriptor

	err := withWorkTreeCacheLock(workTreeCacheDir, func() error {
		writeArchiveRes, err := writeArchive(out, gitDir, workTreeCacheDir, false, opts)
		res = writeArchiveRes
		return err
	})

	return res, err
}

func debugArchive() bool {
	return os.Getenv("WERF_TRUE_GIT_DEBUG_ARCHIVE") == "1"
}

func writeArchive(out io.Writer, gitDir, workTreeCacheDir string, withSubmodules bool, opts ArchiveOptions) (*ArchiveDescriptor, error) {
	var err error

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return nil, fmt.Errorf("bad git dir %s: %s", gitDir, err)
	}

	workTreeCacheDir, err = filepath.Abs(workTreeCacheDir)
	if err != nil {
		return nil, fmt.Errorf("bad work tree cache dir %s: %s", workTreeCacheDir, err)
	}

	if withSubmodules {
		err := checkSubmoduleConstraint()
		if err != nil {
			return nil, err
		}
	}

	workTreeDir, err := prepareWorkTree(gitDir, workTreeCacheDir, opts.Commit, withSubmodules)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare work tree in cache %s for commit %s: %s", workTreeCacheDir, opts.Commit, err)
	}

	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	if err != nil {
		return nil, fmt.Errorf("git open failed: %s", err)
	}

	desc := &ArchiveDescriptor{
		IsEmpty: true,
	}

	absBasePath := filepath.Join(workTreeDir, opts.PathMatcher.BaseFilepath())
	exist, err := util.FileExists(absBasePath)
	if err != nil {
		return nil, fmt.Errorf("file exists %s failed: %s", absBasePath, err)
	}

	if !exist {
		return nil, fmt.Errorf("base path %s entry not found repo", opts.PathMatcher.BaseFilepath())
	}

	info, err := os.Lstat(absBasePath)
	if err != nil {
		return nil, fmt.Errorf("lstat %s failed: %s", absBasePath, err)
	}

	if info.IsDir() {
		desc.Type = DirectoryArchive

		if debugArchive() {
			logboek.Debug.LogF("Found BasePath %s directory: directory archive type\n", absBasePath)
		}
	} else {
		desc.Type = FileArchive

		if debugArchive() {
			logboek.Debug.LogF("Found BasePath %s file: file archive\n", absBasePath)
		}
	}

	tw := tar.NewWriter(out)

	logProcessMsg := fmt.Sprintf("ls-tree (%s)", opts.PathMatcher.String())
	logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
	result, err := ls_tree.LsTree(repository, opts.Commit, opts.PathMatcher, true)
	if err != nil {
		logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		return nil, err
	}
	logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

	logProcessMsg = fmt.Sprintf("ls-tree result walk (%s)", opts.PathMatcher.String())
	logboek.Debug.LogProcessStart(logProcessMsg, logboek.LevelLogProcessStartOptions{})
	if err := result.Walk(func(lsTreeEntry *ls_tree.LsTreeEntry) error {
		logboek.Debug.LogF("ls-tree entry %s\n", lsTreeEntry.FullFilepath)

		desc.IsEmpty = false

		gitFileMode := lsTreeEntry.Mode
		absFilepath := filepath.Join(workTreeDir, lsTreeEntry.FullFilepath)
		relToBasePathFilepath := opts.PathMatcher.TrimFileBaseFilepath(lsTreeEntry.FullFilepath)
		tarEntryName := filepath.ToSlash(relToBasePathFilepath)
		info, err := os.Lstat(absFilepath)
		if err != nil {
			return fmt.Errorf("lstat %s failed: %s", absFilepath, err)
		}

		switch gitFileMode {
		case filemode.Regular, filemode.Executable, filemode.Deprecated:
			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Name:       tarEntryName,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar header for file %s: %s", tarEntryName, err)
			}

			f, err := os.Open(absFilepath)
			if err != nil {
				return fmt.Errorf("unable to open file %s: %s", absFilepath, err)
			}

			_, err = io.Copy(tw, f)
			if err != nil {
				return fmt.Errorf("unable to write data to tar archive from file %s: %s", absFilepath, err)
			}

			err = f.Close()
			if err != nil {
				return fmt.Errorf("error closing file %s: %s", absFilepath, err)
			}

			if debugArchive() {
				logboek.Debug.LogF("Added archive file '%s'\n", relToBasePathFilepath)
			}
		case filemode.Symlink:
			isSymlink := info.Mode()&os.ModeSymlink != 0

			var linkname string
			if isSymlink {
				linkname, err = os.Readlink(absFilepath)
				if err != nil {
					return fmt.Errorf("cannot read symlink %s: %s", absFilepath, err)
				}
			} else {
				data, err := ioutil.ReadFile(absFilepath)
				if err != nil {
					return fmt.Errorf("cannot read file %s: %s", absFilepath, err)
				}

				linkname = string(bytes.TrimSpace(data))
			}

			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       tarEntryName,
				Linkname:   linkname,
				Mode:       int64(gitFileMode),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header for file %s: %s", tarEntryName, err)
			}

			if debugArchive() {
				logboek.Debug.LogF("Added archive symlink %s -> %s\n", relToBasePathFilepath, linkname)
			}

			return nil
		default:
			panic(fmt.Sprintf("unexpected git file mode %s", gitFileMode.String()))
		}

		return nil
	}); err != nil {
		logboek.Debug.LogProcessFail(logboek.LevelLogProcessFailOptions{})
		return nil, err
	}
	logboek.Debug.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot write tar archive: %s", err)
	}

	return desc, nil
}
