package true_git

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/samber/lo"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v3/pkg/tmp_manager"
	"github.com/werf/werf/v3/pkg/true_git/ls_tree"
)

// LFSCredentials are offered to Git LFS only for the scheme and host of URL after
// Git URL rewriting, matching native Git LFS credential scoping.
var (
	_ fmt.Stringer   = LFSCredentials{}
	_ fmt.GoStringer = LFSCredentials{}
)

type LFSCredentials struct {
	URL      string
	Username string
	Password string
}

func (c LFSCredentials) String() string {
	return "<redacted>"
}

func (c LFSCredentials) GoString() string {
	return "true_git.LFSCredentials" + c.String()
}

const (
	lfsPointerMaxSize       = 1024
	lfsCredentialsUserEnv   = "WERF_GIT_LFS_USERNAME"
	lfsCredentialsSecretEnv = "WERF_GIT_LFS_PASSWORD"
)

var (
	lfsPointerVersions     = []string{"https://git-lfs.github.com/spec/v1", "https://hawser.github.com/spec/v1", "http://git-media.io/v/2"}
	lfsPointerKeys         = []string{"version", "oid", "size"}
	lfsPointerHeaderRegexp = regexp.MustCompile("git-media|hawser|git-lfs")
	lfsPointerExtRegexp    = regexp.MustCompile(`^ext-\d{1}-\w+`)
	lfsPointerOidRegexp    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type lfsPointer struct {
	oid  string
	size int64
}

type lfsFile struct {
	path string
	size int64
}

// parseLFSPointer mirrors git-lfs DecodeFrom: content git-lfs would not decode is not a pointer
// and is smudged to itself, so it is returned with ok=false and no error.
func parseLFSPointer(data []byte) (lfsPointer, bool, error) {
	data = bytes.TrimSpace(data)
	if !lfsPointerHeaderRegexp.Match(data) {
		return lfsPointer{}, false, nil
	}

	values := map[string]string{}
	var hasExtensions bool
	line := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		key, value, found := strings.Cut(text, " ")
		if !found || line >= len(lfsPointerKeys) {
			return lfsPointer{}, false, nil
		}
		if key != lfsPointerKeys[line] {
			if !lfsPointerExtRegexp.MatchString(key) {
				return lfsPointer{}, false, nil
			}
			hasExtensions = true
			continue
		}
		values[key] = value
		line++
	}
	if scanner.Err() != nil {
		return lfsPointer{}, false, nil
	}

	if !lo.Contains(lfsPointerVersions, values["version"]) {
		return lfsPointer{}, false, nil
	}
	oid, found := strings.CutPrefix(values["oid"], "sha256:")
	if !found || !lfsPointerOidRegexp.MatchString(oid) {
		return lfsPointer{}, false, nil
	}
	size, err := strconv.ParseInt(values["size"], 10, 64)
	if err != nil || size < 0 {
		return lfsPointer{}, false, nil
	}
	if hasExtensions {
		return lfsPointer{}, false, errors.New("Git LFS pointer extensions are not supported")
	}

	return lfsPointer{oid: oid, size: size}, true, nil
}

// lfsCheckoutOptions keep LFS pointers as committed while the service work tree is checked out:
// the lfs driver is disabled and any other driver running git-lfs skips every path.
func lfsCheckoutOptions() []string {
	return []string{"-c", "filter.lfs.process=", "-c", "filter.lfs.smudge=", "-c", "filter.lfs.required=false", "-c", "lfs.fetchexclude=*"}
}

type lfsCandidate struct {
	path    string
	data    []byte
	pointer lfsPointer
}

// materializeLFSFiles resolves the selected LFS pointers of the commit checked out in
// workTreeDir into temporary files verified against the pointer size and SHA-256.
func materializeLFSFiles(ctx context.Context, repoHandle repo_handle.Handle, result *ls_tree.Result, workTreeDir string, credentials *LFSCredentials) (map[string]lfsFile, func(), error) {
	noop := func() {}

	var candidates []lfsCandidate
	if err := result.Walk(func(entry *ls_tree.LsTreeEntry) error {
		if entry.Mode != filemode.Regular && entry.Mode != filemode.Executable && entry.Mode != filemode.Deprecated {
			return nil
		}
		path := filepath.ToSlash(entry.FullFilepath)
		entryHandle, submodulePath := lfsEntryHandle(repoHandle, path)

		blobSize, err := entryHandle.Repository().Storer.EncodedObjectSize(entry.Hash)
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			// EncodedObjectSize does not follow Git object alternates.
			blob, blobErr := entryHandle.Repository().BlobObject(entry.Hash)
			if blobErr != nil {
				return fmt.Errorf("read Git file %q: %w", path, blobErr)
			}
			blobSize, err = blob.Size, nil
		}
		if err != nil {
			return fmt.Errorf("read Git file size %q: %w", path, err)
		}
		if blobSize > lfsPointerMaxSize {
			return nil
		}

		data, err := entryHandle.ReadBlobObjectContent(entry.Hash)
		if err != nil {
			return fmt.Errorf("read Git file %q: %w", path, err)
		}
		pointer, ok, err := parseLFSPointer(data)
		if err != nil {
			return fmt.Errorf("read Git LFS pointer %q: %w", path, err)
		}
		if !ok {
			return nil
		}
		if submodulePath != "" {
			return fmt.Errorf("Git LFS pointer %q in submodule %q is not supported", path, submodulePath)
		}

		candidates = append(candidates, lfsCandidate{path: path, data: data, pointer: pointer})
		return nil
	}); err != nil {
		return nil, noop, err
	}
	if len(candidates) == 0 {
		return nil, noop, nil
	}

	if _, err := exec.LookPath("git-lfs"); err != nil {
		return nil, noop, fmt.Errorf("git-lfs is required to export Git LFS files: %w", err)
	}

	tmpDir, err := tmp_manager.TempDir("git-lfs-")
	if err != nil {
		return nil, noop, fmt.Errorf("create Git LFS export directory: %w", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			logboek.Context(ctx).Warn().LogF("Remove Git LFS export directory %q: %s\n", tmpDir, err)
		}
	}
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		cleanup()
		return nil, noop, fmt.Errorf("resolve Git LFS export directory: %w", err)
	}

	gitOptions, env, err := lfsSmudgeGitOptionsAndEnv(ctx, workTreeDir, credentials)
	if err != nil {
		cleanup()
		return nil, noop, err
	}
	files := map[string]lfsFile{}
	sizeByOid := map[string]int64{}
	for _, candidate := range candidates {
		output := filepath.Join(absTmpDir, candidate.pointer.oid)
		if size, ok := sizeByOid[candidate.pointer.oid]; ok {
			if size != candidate.pointer.size {
				cleanup()
				return nil, noop, fmt.Errorf("materialize Git LFS file %q: pointer size %d does not match size %d of the same object", candidate.path, candidate.pointer.size, size)
			}
		} else {
			if err := smudgeLFSFile(ctx, workTreeDir, gitOptions, env, candidate, output); err != nil {
				cleanup()
				return nil, noop, fmt.Errorf("materialize Git LFS file %q: %w", candidate.path, err)
			}
			sizeByOid[candidate.pointer.oid] = candidate.pointer.size
		}
		files[candidate.path] = lfsFile{path: output, size: candidate.pointer.size}
	}

	return files, cleanup, nil
}

func lfsEntryHandle(handle repo_handle.Handle, path string) (repo_handle.Handle, string) {
	for _, submodule := range handle.Submodules() {
		submodulePath := filepath.ToSlash(submodule.Config().Path)
		if remaining, ok := strings.CutPrefix(path, submodulePath+"/"); ok {
			entryHandle, nestedPath := lfsEntryHandle(submodule, remaining)
			if nestedPath != "" {
				return entryHandle, submodulePath + "/" + nestedPath
			}
			return entryHandle, submodulePath
		}
	}
	return handle, ""
}

func lfsSmudgeGitOptionsAndEnv(ctx context.Context, workTreeDir string, credentials *LFSCredentials) ([]string, []string, error) {
	// Selected files must not be skipped by host or .lfsconfig fetch filters: the pointer is verified below.
	gitOptions := []string{"-c", "lfs.fetchinclude=", "-c", "lfs.fetchexclude="}
	env := []string{"GIT_LFS_SKIP_SMUDGE=0"}
	if credentials == nil {
		return gitOptions, env, nil
	}

	// Resolve insteadOf with Git itself, without logging URL userinfo from either input or output.
	cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir}, "ls-remote", "--get-url", "--", credentials.URL)
	cmd.Stdout = cmd.OutBuf
	cmd.Stderr = io.Discard
	if err := cmd.Cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("resolve Git LFS credential origin: %w", err)
	}
	origin, err := url.Parse(strings.TrimSpace(cmd.OutBuf.String()))
	if err != nil || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Host == "" {
		return gitOptions, env, nil
	}
	// The secret is passed through the environment of this process only, never through arguments.
	helper := fmt.Sprintf(`!f() { test "$1" = get || return 0; printf 'username=%%s\npassword=%%s\n' "$%s" "$%s"; }; f`, lfsCredentialsUserEnv, lfsCredentialsSecretEnv)
	gitOptions = append(gitOptions, "-c", fmt.Sprintf("credential.%s://%s.helper=%s", origin.Scheme, origin.Host, helper))
	env = append(env, lfsCredentialsUserEnv+"="+credentials.Username, lfsCredentialsSecretEnv+"="+credentials.Password)
	return gitOptions, env, nil
}

func smudgeLFSFile(ctx context.Context, workTreeDir string, gitOptions, env []string, candidate lfsCandidate, output string) error {
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create Git LFS export: %w", err)
	}

	hash := sha256.New()
	cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: workTreeDir, Env: env}, slices.Concat(gitOptions, []string{"lfs", "smudge", "--", candidate.path})...)
	cmd.Stdin = bytes.NewReader(candidate.data)
	// Content must not be buffered in OutBuf/OutErrBuf: it can be arbitrarily large.
	cmd.Stdout = io.MultiWriter(f, hash)
	runErr := cmd.Run(ctx)
	size, seekErr := f.Seek(0, io.SeekCurrent)
	if err := errors.Join(runErr, seekErr, f.Close()); err != nil {
		return fmt.Errorf("run git lfs smudge: %w", err)
	}

	if size != candidate.pointer.size || hex.EncodeToString(hash.Sum(nil)) != candidate.pointer.oid {
		return fmt.Errorf("git lfs smudge output does not match pointer size or SHA-256: got %d bytes with sha256 %x, want %d bytes with sha256 %s", size, hash.Sum(nil), candidate.pointer.size, candidate.pointer.oid)
	}
	return nil
}
