package true_git

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/filemode"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/util"
)

type ArchiveOptions struct {
	Commit      string
	PathMatcher PathMatcher
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

	fileModesFromGit, err := gitWorkTreeFilesModes(gitDir, workTreeDir, withSubmodules)
	if err != nil {
		return nil, err
	}

	desc := &ArchiveDescriptor{
		IsEmpty: true,
	}

	tw := tar.NewWriter(out)

	err = filepath.Walk(workTreeDir, func(absPath string, info os.FileInfo, accessErr error) error {
		if accessErr != nil {
			return fmt.Errorf("error accessing %s: %s", absPath, accessErr)
		}

		baseName := filepath.Base(absPath)
		for _, p := range []string{".git"} {
			if baseName == p {
				return nil
			}
		}

		var relPath string
		if absPath == workTreeDir {
			relPath = "."
		} else {
			relPath = rel(absPath, workTreeDir)
		}

		if relPath == opts.PathMatcher.BasePath() || relPath == "." && opts.PathMatcher.BasePath() == "" {
			if info.IsDir() {
				desc.Type = DirectoryArchive

				if debugArchive() {
					fmt.Printf("Found BasePath %s directory: directory archive type\n", relPath)
				}
			} else {
				desc.Type = FileArchive

				if debugArchive() {
					fmt.Printf("Found BasePath %s file: file archive\n", relPath)
				}
			}
		}

		if info.IsDir() {
			return nil
		}

		if !opts.PathMatcher.MatchPath(relPath) {
			if debugArchive() {
				fmt.Printf("Excluded path %s by path filter %s\n", relPath, opts.PathMatcher.String())
			}
			return nil
		}

		unixRelPath := util.ToLinuxContainerPath(relPath)
		fileModeFromGit, exist := fileModesFromGit[unixRelPath]
		if !exist {
			return fmt.Errorf("cannot process filename %s", unixRelPath)
		}

		tarEntryName := util.ToLinuxContainerPath(opts.PathMatcher.TrimFileBasePath(relPath))

		desc.IsEmpty = false

		if fileModeFromGit == filemode.Symlink {
			isSymlink := info.Mode()&os.ModeSymlink != 0

			var linkname string
			if isSymlink {
				linkname, err = os.Readlink(absPath)
				if err != nil {
					return fmt.Errorf("cannot read symlink %s: %s", absPath, err)
				}
			} else {
				data, err := ioutil.ReadFile(absPath)
				if err != nil {
					return fmt.Errorf("cannot read file %s: %s", absPath, err)
				}

				linkname = string(bytes.TrimSpace(data))
			}

			err = tw.WriteHeader(&tar.Header{
				Format:     tar.FormatGNU,
				Typeflag:   tar.TypeSymlink,
				Name:       tarEntryName,
				Linkname:   linkname,
				Mode:       int64(fileModeFromGit),
				Size:       info.Size(),
				ModTime:    info.ModTime(),
				AccessTime: info.ModTime(),
				ChangeTime: info.ModTime(),
			})
			if err != nil {
				return fmt.Errorf("unable to write tar symlink header for file %s: %s", tarEntryName, err)
			}

			if debugArchive() {
				fmt.Printf("Added archive symlink %s -> %s\n", relPath, linkname)
			}

			return nil
		}

		err = tw.WriteHeader(&tar.Header{
			Format:     tar.FormatGNU,
			Name:       tarEntryName,
			Mode:       int64(fileModeFromGit),
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			AccessTime: info.ModTime(),
			ChangeTime: info.ModTime(),
		})
		if err != nil {
			return fmt.Errorf("unable to write tar header for file %s: %s", tarEntryName, err)
		}

		file, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %s", absPath, err)
		}

		_, err = io.Copy(tw, file)
		if err != nil {
			return fmt.Errorf("unable to write data to tar archive from file %s: %s", relPath, err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("error closing file %s: %s", absPath, err)
		}

		if debugArchive() {
			logboek.LogF("Added archive file '%s'\n", relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("entries iteration failed in %s: %s", workTreeDir, err)
	}

	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot write tar archive: %s", err)
	}

	if desc.Type == "" {
		return nil, fmt.Errorf("base path %s entry not found repo", opts.PathMatcher.BasePath())
	}

	return desc, nil
}

func gitWorkTreeFilesModes(repoDir, workTreeDir string, withSubmodules bool) (map[string]filemode.FileMode, error) {
	modeByRelPath := map[string]filemode.FileMode{}
	var gitLsFilesCommandOutputs []*bytes.Buffer

	execArgs := []string{
		"git", "-c", "core.quotePath=false", "--git-dir", repoDir, "--work-tree", workTreeDir,
		"ls-files", "--stage",
	}

	outBuf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})
	cmd := exec.Command(execArgs[0], execArgs[1:]...)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("`%s` failed: %s\nOut: %s\nErr: %s", strings.Join(execArgs, " "), err,
			outBuf.String(), errBuf.String())
	}

	gitLsFilesCommandOutputs = append(gitLsFilesCommandOutputs, outBuf)

	if withSubmodules {
		execArgs := []string{
			"git", "--git-dir", repoDir, "--work-tree", workTreeDir,
			"submodule", "foreach", "--recursive", "git", "-c", "core.quotePath=false", "ls-files", "--stage",
		}

		outBuf := bytes.NewBuffer([]byte{})
		errBuf := bytes.NewBuffer([]byte{})
		cmd := exec.Command(execArgs[0], execArgs[1:]...)
		cmd.Stdout = outBuf
		cmd.Stderr = errBuf
		cmd.Dir = workTreeDir

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("`%s` failed: %s\nOut: %s\nErr: %s", strings.Join(execArgs, " "), err,
				outBuf.String(), errBuf.String())
		}

		gitLsFilesCommandOutputs = append(gitLsFilesCommandOutputs, outBuf)
	}

	submoduleNameRegexp := regexp.MustCompile("Entering '(.*)'$")
	lsFilesLineArgsSplitterRegexp := regexp.MustCompile("[[:space:]]+")

	for _, b := range gitLsFilesCommandOutputs {
		scanner := bufio.NewScanner(b)
		var submodulePath string

		for scanner.Scan() {
			line := scanner.Text()

			if match := submoduleNameRegexp.FindStringSubmatch(line); match != nil {
				submodulePath = match[1]
				continue
			}

			parts := lsFilesLineArgsSplitterRegexp.Split(line, 4)
			if len(parts) != 4 {
				panic(fmt.Sprintf("unexpected `git ls files` line format `%s`", line))
			}

			modeStr := parts[0]
			relFilePath := path.Join(submodulePath, parts[3])

			fileMode, err := filemode.New(modeStr)
			if err != nil {
				panic(fmt.Sprintf("unexpected `git ls files` line format `%s`", line))
			}

			modeByRelPath[relFilePath] = fileMode
		}
	}

	return modeByRelPath, nil
}

func rel(targetPath, basePath string) string {
	if basePath == "" {
		return targetPath
	}

	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		panic(err)
	}

	return relPath
}
