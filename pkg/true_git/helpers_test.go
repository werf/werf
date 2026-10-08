package true_git

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/werf/v3/pkg/git_repo/repo_handle"
	"github.com/werf/werf/v3/test/pkg/utils"
)

type blockingGitFilter struct {
	command     string
	armPath     string
	startedPath string
	proceedPath string
}

func newBlockingGitFilter(dir, name string) blockingGitFilter {
	filter := blockingGitFilter{
		command:     filepath.Join(dir, name+".sh"),
		armPath:     filepath.Join(dir, name+"-arm"),
		startedPath: filepath.Join(dir, name+"-started"),
		proceedPath: filepath.Join(dir, name+"-proceed"),
	}
	script := fmt.Sprintf(
		"#!/bin/sh\nif [ -f %q ]; then\n\ttouch %q\n\tfor i in $(seq 1 600); do [ -f %q ] && break; sleep 0.05; done\nfi\ncat\n",
		filter.armPath, filter.startedPath, filter.proceedPath,
	)
	Expect(os.WriteFile(filter.command, []byte(script), 0o755)).To(Succeed())
	return filter
}

func (f blockingGitFilter) arm() {
	Expect(os.WriteFile(f.armPath, []byte("arm"), 0o644)).To(Succeed())
}

func (f blockingGitFilter) terminateWhenBlocked(ctx context.Context) (context.Context, chan struct{}) {
	terminationCtx := graceful.WithTermination(ctx)
	done := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		defer close(done)
		for i := 0; i < 600; i++ {
			if _, err := os.Stat(f.startedPath); err == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		graceful.Terminate(terminationCtx, fmt.Errorf("sibling task failed"), 1)
		<-terminationCtx.Done()
		Expect(os.WriteFile(f.proceedPath, []byte("go"), 0o644)).To(Succeed())
	}()
	return terminationCtx, done
}

// gitCommitSucceed pins identity and disables signing so commits do not depend on ambient
// global git config: a developer's commit.gpgsign needs a working gpg-agent, which
// intermittently fails under Ginkgo's parallel procs, and specs that isolate themselves via
// GIT_CONFIG_GLOBAL lose the only user.name/user.email the CI runner has.
func gitCommitSucceed(ctx context.Context, workTreeDir string, args ...string) {
	utils.RunSucceedCommand(ctx, workTreeDir, "git", append([]string{
		"-c", "commit.gpgsign=false",
		"-c", "user.email=werf@werf.io",
		"-c", "user.name=werf",
		"commit",
	}, args...)...)
}

func gitSucceed(ctx context.Context, dir string, args ...string) string {
	return utils.SucceedCommandOutputString(ctx, dir, "git", append([]string{
		"-c", "commit.gpgsign=false",
		"-c", "user.email=werf@werf.io",
		"-c", "user.name=werf",
	}, args...)...)
}

func gitSucceedTrimmed(ctx context.Context, dir string, args ...string) string {
	return strings.TrimSpace(gitSucceed(ctx, dir, args...))
}

func gitSucceedWithStdin(ctx context.Context, dir, stdin string, args ...string) string {
	out, err := utils.RunCommandWithOptions(ctx, dir, "git", args, utils.RunCommandOptions{
		ToStdin:       stdin,
		ShouldSucceed: true,
		NoStderr:      true,
	})
	Expect(err).ToNot(HaveOccurred())
	return strings.TrimSpace(string(out))
}

func gitInitRepo(ctx context.Context, dir string) {
	utils.MkdirAll(dir)
	gitSucceed(ctx, dir, "-c", "init.defaultBranch=main", "init")
	gitSucceed(ctx, dir, "commit", "--allow-empty", "-m", "init")
}

// setEnvForSpec sets an environment variable for the duration of the current spec. Specs of one
// Ginkgo process run serially, so a process-wide variable restored on teardown stays contained.
func setEnvForSpec(key, value string) {
	previous, had := os.LookupEnv(key)
	Expect(os.Setenv(key, value)).To(Succeed())
	DeferCleanup(func() {
		if had {
			Expect(os.Setenv(key, previous)).To(Succeed())
			return
		}
		Expect(os.Unsetenv(key)).To(Succeed())
	})
}

// isolateGitConfig detaches the spec from the developer's global and system git config, so an
// ambient setting (notably protocol.file.allow=always) cannot mask a missing production option.
func isolateGitConfig() {
	setEnvForSpec("GIT_CONFIG_GLOBAL", os.DevNull)
	setEnvForSpec("GIT_CONFIG_SYSTEM", os.DevNull)
	setEnvForSpec("GIT_TERMINAL_PROMPT", "0")
}

func gitAddSubmoduleSucceed(ctx context.Context, superRepoDir, url, path string) {
	gitSucceed(ctx, superRepoDir, "-c", "protocol.file.allow=always", "submodule", "add", url, path)
}

func gitUpdateSubmodulesSucceed(ctx context.Context, repoDir string, args ...string) {
	gitSucceed(ctx, repoDir, append([]string{"-c", "protocol.file.allow=always", "submodule", "update"}, args...)...)
}

func gitInitRepoWithFile(ctx context.Context, dir, fileName, content string) {
	gitInitRepo(ctx, dir)
	Expect(os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)).To(Succeed())
	gitSucceed(ctx, dir, "add", ".")
	gitSucceed(ctx, dir, "commit", "-m", "content")
}

func expectSubmoduleFile(gitDir, workTreeDir, fileName, content string, submodulePaths ...string) {
	GinkgoHelper()
	repository, err := GitOpenWithCustomWorktreeDir(gitDir, workTreeDir)
	Expect(err).ToNot(HaveOccurred())
	handle, err := repo_handle.NewHandle(repository)
	Expect(err).ToNot(HaveOccurred())
	for _, path := range submodulePaths {
		submodule, err := handle.Submodule(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(submodule.Status().Current).To(Equal(submodule.Status().Expected))
		handle = submodule
	}
	head, err := handle.Repository().Head()
	Expect(err).ToNot(HaveOccurred())
	tree, err := handle.GetCommitTree(head.Hash())
	Expect(err).ToNot(HaveOccurred())
	entry, err := tree.FindEntry(fileName)
	Expect(err).ToNot(HaveOccurred())
	data, err := handle.ReadBlobObjectContent(entry.Hash)
	Expect(err).ToNot(HaveOccurred())
	Expect(string(data)).To(Equal(content))
}

func lfsTestPointer(content string) string {
	return fmt.Sprintf("version https://git-lfs.github.com/spec/v1\noid sha256:%x\nsize %d\n", sha256.Sum256([]byte(content)), len(content))
}

func writeLFSTestFile(repoDir, name, content string, objectPresent bool) {
	file := filepath.Join(repoDir, name)
	Expect(os.MkdirAll(filepath.Dir(file), 0o755)).To(Succeed())
	Expect(os.WriteFile(file, []byte(lfsTestPointer(content)), 0o755)).To(Succeed())
	if objectPresent {
		oid := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		object := filepath.Join(repoDir, ".git", "lfs", "objects", oid[:2], oid[2:4], oid)
		Expect(os.MkdirAll(filepath.Dir(object), 0o755)).To(Succeed())
		Expect(os.WriteFile(object, []byte(content), 0o644)).To(Succeed())
	}
}

func readTestTar(data []byte) map[string]string {
	files := make(map[string]string)
	reader := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return files
		}
		Expect(err).NotTo(HaveOccurred())
		if header.Typeflag != tar.TypeReg {
			continue
		}
		content, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		files[header.Name] = string(content)
	}
}

func nativeLFSTestSmudge(ctx context.Context, repoDir, path, pointer string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "lfs", "smudge", "--", path)
	cmd.Dir = repoDir
	cmd.Stdin = strings.NewReader(pointer)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return string(output), fmt.Errorf("native LFS smudge: %w: %s", err, stderr.String())
	}
	return string(output), nil
}

type lfsTestServer struct {
	server                  *httptest.Server
	missingRequests         atomic.Int32
	authorizedRequests      atomic.Int32
	unexpectedAuthorization atomic.Bool
	downloads               atomic.Int32
	redirect                bool
}

func newLFSTestServer(content, authorization, downloadURL string) *lfsTestServer {
	fixture := &lfsTestServer{}
	oid := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		if r.Method == http.MethodPost && r.URL.Path == "/repo.git/info/lfs/objects/batch" {
			if authorization != "" && r.Header.Get("Authorization") != authorization {
				if r.Header.Get("Authorization") != "" {
					fixture.unexpectedAuthorization.Store(true)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Header().Set("WWW-Authenticate", `Basic realm="lfs-test"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if authorization != "" {
				fixture.authorizedRequests.Add(1)
			}
			var batch struct {
				Operation string `json:"operation"`
				Objects   []struct {
					OID  string `json:"oid"`
					Size int64  `json:"size"`
				} `json:"objects"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&batch)).To(Succeed())
			Expect(batch.Operation).To(Equal("download"))
			Expect(batch.Objects).To(HaveLen(1))
			object := batch.Objects[0]
			w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
			if object.OID != oid {
				fixture.missingRequests.Add(1)
				Expect(json.NewEncoder(w).Encode(map[string]any{"objects": []any{map[string]any{
					"oid": object.OID, "size": object.Size, "error": map[string]any{"code": 404, "message": "fixture object unavailable"},
				}}})).To(Succeed())
				return
			}
			Expect(object.Size).To(Equal(int64(len(content))))
			href := downloadURL
			if href == "" || fixture.redirect {
				href = fixture.server.URL + "/object"
			}
			Expect(json.NewEncoder(w).Encode(map[string]any{"transfer": "basic", "objects": []any{map[string]any{
				"oid": oid, "size": len(content), "actions": map[string]any{"download": map[string]any{"href": href}},
			}}})).To(Succeed())
			return
		}
		Expect(r.Method).To(Equal(http.MethodGet))
		Expect(r.URL.Path).To(Equal("/object"))
		if fixture.redirect {
			http.Redirect(w, r, downloadURL, http.StatusFound)
			return
		}
		fixture.downloads.Add(1)
		_, err := w.Write([]byte(content))
		Expect(err).NotTo(HaveOccurred())
	}))
	DeferCleanup(fixture.server.Close)
	return fixture
}

func configureLFSTestCredentials(ctx context.Context, repoDir, endpoint, source string) {
	parsed, err := url.Parse(endpoint)
	Expect(err).NotTo(HaveOccurred())
	helper := filepath.Join(SuiteData.TestDirPath, "credential-helper")
	script := fmt.Sprintf("#!/bin/sh\nhost=\nwhile IFS= read -r line; do\n case \"$line\" in host=*) host=${line#host=} ;; esac\ndone\nif [ \"$1\" = get ] && [ \"$host\" = %q ]; then\n printf 'username=lfs-user\\npassword=dummy-lfs-secret\\n'\nfi\n", parsed.Host)
	Expect(os.WriteFile(helper, []byte(script), 0o755)).To(Succeed())
	args := []string{"config"}
	if source == "global" {
		config := filepath.Join(SuiteData.TestDirPath, "credential.gitconfig")
		setEnvForSpec("GIT_CONFIG_GLOBAL", config)
		args = append(args, "--file", config)
	}
	gitSucceed(ctx, repoDir, append(args, "credential.helper", "")...)
	gitSucceed(ctx, repoDir, append(args, "credential."+endpoint+".helper", helper)...)
	gitSucceed(ctx, repoDir, append(args, "credential.useHttpPath", "true")...)
}
