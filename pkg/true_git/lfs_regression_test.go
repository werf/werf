package true_git

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/path_matcher"
	"github.com/werf/werf/v3/test/pkg/suite_init"
)

var _ = ginkgo.Describe("LFS public entrypoint regressions", func() {
	suite_init.NewWerfInitData(SuiteData.TmpDirData)
	var repoDir, cacheDir string
	const content = "selected native LFS contents\x00\xff"

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		isolateGitConfig()
		setEnvForSpec("GIT_TERMINAL_PROMPT", "0")
		setEnvForSpec("GIT_ASKPASS", "false")
		setEnvForSpec("SSH_ASKPASS", "false")
		setEnvForSpec("GIT_CONFIG_COUNT", "0")
		setEnvForSpec("GIT_CONFIG_PARAMETERS", "")
		setEnvForSpec("GIT_LFS_SKIP_SMUDGE", "0")
		setEnvForSpec("GIT_LFS_SKIP_DOWNLOAD_ERRORS", "0")
		setEnvForSpec("HOME", SuiteData.TestDirPath)
		setEnvForSpec("XDG_CONFIG_HOME", SuiteData.TestDirPath)
		setEnvForSpec("NETRC", filepath.Join(SuiteData.TestDirPath, "absent-netrc"))
		_, err := exec.LookPath("git-lfs")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		repoDir = filepath.Join(SuiteData.TestDirPath, "repo")
		cacheDir = filepath.Join(SuiteData.TestDirPath, "archive-cache")
		gitInitRepo(ctx, repoDir)
		gomega.Expect(Init(ctx, Options{})).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("never exports native-accepted pointers as successful text",
		func(ctx ginkgo.SpecContext, version, prefix, newline string) {
			writeLFSTestFile(repoDir, "payload.bin", content, true)
			pointer := prefix + strings.ReplaceAll(strings.Replace(lfsTestPointer(content), "https://git-lfs.github.com/spec/v1", version, 1), "\n", newline)
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(pointer), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "lfs", "pointer", "--check", "--file=payload.bin")
			gitSucceed(ctx, repoDir, "add", "payload.bin")
			gitCommitSucceed(ctx, repoDir, "-m", "native pointer variant")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+repoDir)
			for _, checksum := range []string{"", "dockerfile-context"} {
				var output bytes.Buffer
				err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
					Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin", ContentChecksum: checksum,
					PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
			}
		},
		ginkgo.Entry("canonical", "https://git-lfs.github.com/spec/v1", "", "\n"),
		ginkgo.Entry("legacy git-media", "http://git-media.io/v/2", "", "\n"),
		ginkgo.Entry("legacy hawser", "https://hawser.github.com/spec/v1", "", "\n"),
		ginkgo.Entry("leading LF", "https://git-lfs.github.com/spec/v1", "\n", "\n"),
		ginkgo.Entry("leading space", "https://git-lfs.github.com/spec/v1", " ", "\n"),
		ginkgo.Entry("CRLF", "https://git-lfs.github.com/spec/v1", "", "\r\n"),
	)

	ginkgo.DescribeTable("preserves native LFS clean outside export preparation",
		func(ctx ginkgo.SpecContext, operation string) {
			gitSucceed(ctx, repoDir, "lfs", "install", "--local", "--skip-repo")
			gitSucceed(ctx, repoDir, "config", "filter.lfs.clean", "")
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("*.bin filter=lfs diff=lfs merge=lfs -text\n"), 0o644)).To(gomega.Succeed())
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(content), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "native LFS clean")
			gomega.Expect(gitSucceed(ctx, repoDir, "show", "HEAD:payload.bin")).To(gomega.Equal(lfsTestPointer(content)))
			commit := gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")
			gitSucceed(ctx, repoDir, "config", "user.name", "LFS test")
			gitSucceed(ctx, repoDir, "config", "user.email", "lfs@example.invalid")
			indexBefore, err := os.ReadFile(filepath.Join(repoDir, ".git", "index"))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			updated := content + " dev update"
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(updated), 0o644)).To(gomega.Succeed())
			if operation == "add" {
				cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: repoDir}, "add", "payload.bin")
				gomega.Expect(cmd.Run(ctx)).To(gomega.Succeed())
				gomega.Expect(gitSucceed(ctx, repoDir, "show", ":payload.bin")).To(gomega.Equal(lfsTestPointer(updated)))
				return
			}
			serviceCommit, err := SyncSourceWorktreeWithServiceBranch(ctx, filepath.Join(repoDir, ".git"), repoDir, cacheDir, commit,
				SyncSourceWorktreeWithServiceBranchOptions{ServiceBranch: "_werf-dev-lfs"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(gitSucceed(ctx, repoDir, "show", serviceCommit+":payload.bin")).To(gomega.Equal(lfsTestPointer(updated)))
			gomega.Expect(os.ReadFile(filepath.Join(repoDir, ".git", "index"))).To(gomega.Equal(indexBefore))
			gomega.Expect(gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD")).To(gomega.Equal(commit))
			gomega.Expect(os.ReadFile(filepath.Join(repoDir, "payload.bin"))).To(gomega.Equal([]byte(updated)))
		},
		ginkgo.Entry("add", "add"),
		ginkgo.Entry("service add", "service"),
	)

	ginkgo.DescribeTable("preserves transport configuration and never smudges an excluded alias-LFS object",
		func(ctx ginkgo.SpecContext, configSource, checksum string) {
			server := newLFSTestServer(content, "", "")
			writeLFSTestFile(repoDir, "wanted.bin", content, false)
			writeLFSTestFile(repoDir, "excluded.bin", "unavailable excluded object", false)
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("excluded.bin filter=alias-lfs\n"), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "selected and excluded pointers")
			gitSucceed(ctx, repoDir, "config", "filter.alias-lfs.smudge", "git-lfs smudge -- %f")
			gitSucceed(ctx, repoDir, "config", "filter.alias-lfs.clean", "git-lfs clean -- %f")
			gitSucceed(ctx, repoDir, "config", "filter.alias-lfs.required", "true")
			origin := server.server.URL + "/repo.git"
			if configSource != "direct" {
				origin = "file://" + filepath.Join(SuiteData.TestDirPath, "unavailable-origin-alias")
				key := "url." + server.server.URL + "/repo.git.insteadOf"
				switch configSource {
				case "local":
					gitSucceed(ctx, repoDir, "config", key, origin)
				case "environment":
					setEnvForSpec("GIT_CONFIG_COUNT", "1")
					setEnvForSpec("GIT_CONFIG_KEY_0", key)
					setEnvForSpec("GIT_CONFIG_VALUE_0", origin)
				default:
					config := filepath.Join(SuiteData.TestDirPath, configSource+".gitconfig")
					gitSucceed(ctx, repoDir, "config", "--file", config, key, origin)
					setEnvForSpec("GIT_CONFIG_"+strings.ToUpper(configSource), config)
				}
			}
			gitSucceed(ctx, repoDir, "remote", "add", "origin", origin)
			gomega.Expect(nativeLFSTestSmudge(ctx, repoDir, "wanted.bin", lfsTestPointer(content))).To(gomega.Equal(content))
			gomega.Expect(server.downloads.Load()).To(gomega.BeNumerically(">", 0))
			_, controlErr := nativeLFSTestSmudge(ctx, repoDir, "excluded.bin", lfsTestPointer("unavailable excluded object"))
			gomega.Expect(controlErr).To(gomega.HaveOccurred())
			gomega.Expect(server.missingRequests.Load()).To(gomega.BeNumerically(">", 0), "negative control must contact the same endpoint for an included unavailable object")
			server.missingRequests.Store(0)
			server.downloads.Store(0)
			gomega.Expect(os.RemoveAll(filepath.Join(repoDir, ".git", "lfs", "objects"))).To(gomega.Succeed())
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), ContentChecksum: checksum,
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{IncludeGlobs: []string{"wanted.bin"}}),
			})
			gomega.Expect(server.missingRequests.Load()).To(gomega.BeZero(), "excluded object must not be requested even during checkout")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{"wanted.bin": content}))
			gomega.Expect(server.downloads.Load()).To(gomega.BeNumerically(">", 0))
		},
		ginkgo.Entry("Stapel direct origin", "direct", ""),
		ginkgo.Entry("Dockerfile direct origin", "direct", "context-checksum"),
		ginkgo.Entry("Stapel global rewrite", "global", ""),
		ginkgo.Entry("Dockerfile global rewrite", "global", "context-checksum"),
		ginkgo.Entry("Stapel environment rewrite", "environment", ""),
		ginkgo.Entry("Dockerfile environment rewrite", "environment", "context-checksum"),
		ginkgo.Entry("local rewrite", "local", ""),
		ginkgo.Entry("system rewrite", "system", ""),
	)

	ginkgo.DescribeTable("uses native and explicit credentials without forwarding them to another download host",
		func(ctx ginkgo.SpecContext, credentials, transfer string) {
			var otherRequests atomic.Int32
			var leakedAuthorization atomic.Bool
			other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer ginkgo.GinkgoRecover()
				otherRequests.Add(1)
				if r.Header.Get("Authorization") != "" {
					leakedAuthorization.Store(true)
				}
				_, err := w.Write([]byte(content))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}))
			defer other.Close()
			otherURL := strings.Replace(other.URL, "127.0.0.1", "localhost", 1) + "/object"
			authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte("lfs-user:dummy-lfs-secret"))
			server := newLFSTestServer(content, authorization, otherURL)
			if transfer == "redirect" {
				server.redirect = true
			}
			origin := server.server.URL + "/repo.git"
			if credentials == "userinfo" {
				origin = strings.Replace(origin, "://", "://lfs-user:dummy-lfs-secret@", 1)
			} else {
				source := credentials
				if strings.HasPrefix(credentials, "explicit") {
					source = "local"
				}
				if credentials == "rewrite" || credentials == "explicit-rewrite" {
					source = "global"
					origin = "file://" + filepath.Join(SuiteData.TestDirPath, "authenticated-origin-alias")
					setEnvForSpec("GIT_CONFIG_COUNT", "1")
					setEnvForSpec("GIT_CONFIG_KEY_0", "url."+server.server.URL+"/repo.git.insteadOf")
					setEnvForSpec("GIT_CONFIG_VALUE_0", origin)
				}
				configureLFSTestCredentials(ctx, repoDir, server.server.URL, source)
			}
			writeLFSTestFile(repoDir, "payload.bin", content, false)
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "authenticated pointer")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", origin)
			gomega.Expect(nativeLFSTestSmudge(ctx, repoDir, "payload.bin", lfsTestPointer(content))).To(gomega.Equal(content))
			gomega.Expect(leakedAuthorization.Load()).To(gomega.BeFalse())
			gomega.Expect(server.authorizedRequests.Load()).To(gomega.BeNumerically(">", 0))
			gomega.Expect(otherRequests.Load()).To(gomega.BeNumerically(">", 0))
			server.authorizedRequests.Store(0)
			otherRequests.Store(0)
			gomega.Expect(os.RemoveAll(filepath.Join(repoDir, ".git", "lfs", "objects"))).To(gomega.Succeed())
			var explicit *LFSCredentials
			if strings.HasPrefix(credentials, "explicit") {
				config := []string{"config"}
				if credentials == "explicit-rewrite" {
					config = append(config, "--global")
				}
				gitSucceed(ctx, repoDir, append(config, "--unset-all", "credential."+server.server.URL+".helper")...)
				explicit = &LFSCredentials{URL: origin, Username: "lfs-user", Password: "dummy-lfs-secret"}
			}
			if strings.HasPrefix(credentials, "explicit-stale-") {
				gitSucceed(ctx, repoDir, "config", "--unset-all", "credential.helper")
				staleHelper := `!f() { test "$1" = get || return 0; printf 'username=stale-user\npassword=stale-secret\n'; }; f`
				if credentials == "explicit-stale-global" {
					setEnvForSpec("GIT_CONFIG_GLOBAL", filepath.Join(SuiteData.TestDirPath, "stale-credentials.gitconfig"))
					gitSucceed(ctx, repoDir, "config", "--global", "credential.helper", staleHelper)
				} else {
					setEnvForSpec("GIT_CONFIG_COUNT", "1")
					setEnvForSpec("GIT_CONFIG_KEY_0", "credential.helper")
					setEnvForSpec("GIT_CONFIG_VALUE_0", staleHelper)
				}
				_, controlErr := nativeLFSTestSmudge(ctx, repoDir, "payload.bin", lfsTestPointer(content))
				gomega.Expect(controlErr).To(gomega.HaveOccurred())
				gomega.Expect(server.unexpectedAuthorization.Load()).To(gomega.BeTrue(), "negative control must send stale helper credentials")
				server.unexpectedAuthorization.Store(false)
			}
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin", LFSCredentials: explicit,
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})
			gomega.Expect(leakedAuthorization.Load()).To(gomega.BeFalse(), "origin credentials must not reach another host")
			gomega.Expect(server.unexpectedAuthorization.Load()).To(gomega.BeFalse(), "explicit credentials must take precedence over stale helpers")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
			gomega.Expect(server.authorizedRequests.Load()).To(gomega.BeNumerically(">", 0))
			gomega.Expect(otherRequests.Load()).To(gomega.BeNumerically(">", 0))
		},
		ginkgo.Entry("global helper and separate download host", "global", "direct"),
		ginkgo.Entry("local helper and separate download host", "local", "direct"),
		ginkgo.Entry("global helper and cross-host redirect", "global", "redirect"),
		ginkgo.Entry("helper bound to insteadOf effective origin", "rewrite", "direct"),
		ginkgo.Entry("URL userinfo and separate download host", "userinfo", "direct"),
		ginkgo.Entry("explicit credentials and separate download host", "explicit", "direct"),
		ginkgo.Entry("explicit credentials and cross-host redirect", "explicit", "redirect"),
		ginkgo.Entry("explicit credentials after environment origin rewrite", "explicit-rewrite", "direct"),
		ginkgo.Entry("explicit credentials override a stale global helper", "explicit-stale-global", "direct"),
		ginkgo.Entry("explicit credentials override a stale environment helper", "explicit-stale-environment", "direct"),
	)

	ginkgo.DescribeTable("keeps origin credentials away from a cross-host LFS endpoint requiring authentication",
		func(ctx ginkgo.SpecContext, destinationHelper bool) {
			authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte("lfs-user:dummy-lfs-secret"))
			server := newLFSTestServer(content, authorization, "")
			origin := strings.Replace(server.server.URL, "127.0.0.1", "localhost", 1) + "/repo.git"
			writeLFSTestFile(repoDir, "payload.bin", content, false)
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".lfsconfig"), []byte("[lfs]\nurl = "+server.server.URL+"/repo.git/info/lfs\n"), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "cross-host LFS endpoint")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", origin)
			configureLFSTestCredentials(ctx, repoDir, server.server.URL, "local")
			gomega.Expect(nativeLFSTestSmudge(ctx, repoDir, "payload.bin", lfsTestPointer(content))).To(gomega.Equal(content))
			gomega.Expect(server.authorizedRequests.Load()).To(gomega.BeNumerically(">", 0))
			gomega.Expect(server.unexpectedAuthorization.Load()).To(gomega.BeFalse())
			gomega.Expect(os.RemoveAll(filepath.Join(repoDir, ".git", "lfs", "objects"))).To(gomega.Succeed())
			server.authorizedRequests.Store(0)
			if !destinationHelper {
				gitSucceed(ctx, repoDir, "config", "--unset-all", "credential."+server.server.URL+".helper")
			}
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin",
				LFSCredentials: &LFSCredentials{URL: origin, Username: "origin-user", Password: "origin-secret"},
				PathMatcher:    path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})
			gomega.Expect(server.unexpectedAuthorization.Load()).To(gomega.BeFalse(), "origin credentials must not answer another host's challenge")
			if !destinationHelper {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(server.authorizedRequests.Load()).To(gomega.BeZero())
				return
			}
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
			gomega.Expect(server.authorizedRequests.Load()).To(gomega.BeNumerically(">", 0))
		},
		ginkgo.Entry("destination's own helper", true),
		ginkgo.Entry("missing destination credentials", false),
	)

	ginkgo.DescribeTable("offers explicit credentials across paths on the effective origin host",
		func(ctx ginkgo.SpecContext, transfer string) {
			authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte("lfs-user:dummy-lfs-secret"))
			var authorized atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer ginkgo.GinkgoRecover()
				if r.URL.Path == "/object" {
					_, err := w.Write([]byte(content))
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					return
				}
				gomega.Expect(r.Method).To(gomega.Equal(http.MethodPost))
				if r.Header.Get("Authorization") != authorization {
					w.Header().Set("WWW-Authenticate", `Basic realm="lfs-test"`)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if r.URL.Path == "/repo.git/info/lfs/objects/batch" {
					http.Redirect(w, r, server.URL+"/other/lfs/objects/batch", http.StatusTemporaryRedirect)
					return
				}
				gomega.Expect(r.URL.Path).To(gomega.Equal("/other/lfs/objects/batch"))
				authorized.Add(1)
				w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
				gomega.Expect(json.NewEncoder(w).Encode(map[string]any{"transfer": "basic", "objects": []any{map[string]any{
					"oid": fmt.Sprintf("%x", sha256.Sum256([]byte(content))), "size": len(content),
					"actions": map[string]any{"download": map[string]any{"href": server.URL + "/object"}},
				}}})).To(gomega.Succeed())
			}))
			defer server.Close()
			writeLFSTestFile(repoDir, "payload.bin", content, false)
			if transfer == "lfsconfig" {
				gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".lfsconfig"), []byte("[lfs]\nurl = "+server.URL+"/other/lfs\n"), 0o644)).To(gomega.Succeed())
			}
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "same-host LFS credentials")
			origin := server.URL + "/repo.git"
			gitSucceed(ctx, repoDir, "remote", "add", "origin", origin)
			var output bytes.Buffer
			gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin",
				LFSCredentials: &LFSCredentials{URL: origin, Username: "lfs-user", Password: "dummy-lfs-secret"},
				PathMatcher:    path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})).To(gomega.Succeed())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
			gomega.Expect(authorized.Load()).To(gomega.BeNumerically(">", 0))
		},
		ginkgo.Entry("committed endpoint on another path", "lfsconfig"),
		ginkgo.Entry("same-host batch redirect", "redirect"),
	)

	ginkgo.It("uses committed .lfsconfig as native transport configuration", func(ctx ginkgo.SpecContext) {
		server := newLFSTestServer(content, "", "")
		writeLFSTestFile(repoDir, "payload.bin", content, false)
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".lfsconfig"), []byte("[lfs]\nurl = "+server.server.URL+"/repo.git/info/lfs\n"), 0o644)).To(gomega.Succeed())
		gitSucceed(ctx, repoDir, "add", ".")
		gitCommitSucceed(ctx, repoDir, "-m", "LFS transport endpoint")
		gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+filepath.Join(SuiteData.TestDirPath, "unavailable-origin"))
		gomega.Expect(nativeLFSTestSmudge(ctx, repoDir, "payload.bin", lfsTestPointer(content))).To(gomega.Equal(content))
		gomega.Expect(server.downloads.Load()).To(gomega.BeNumerically(">", 0))
		server.downloads.Store(0)
		gomega.Expect(os.RemoveAll(filepath.Join(repoDir, ".git", "lfs", "objects"))).To(gomega.Succeed())
		var output bytes.Buffer
		gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
		gomega.Expect(server.downloads.Load()).To(gomega.BeNumerically(">", 0))
	})

	ginkgo.It("exports ordinary submodules with inherited transport configuration", func(ctx ginkgo.SpecContext) {
		subDir := filepath.Join(SuiteData.TestDirPath, "sub-origin")
		gitInitRepoWithFile(ctx, subDir, "plain.txt", "ordinary submodule bytes")
		gitAddSubmoduleSucceed(ctx, repoDir, subDir, "sub")
		gitCommitSucceed(ctx, repoDir, "-m", "ordinary submodule")
		setEnvForSpec("GIT_CONFIG_COUNT", "1")
		setEnvForSpec("GIT_CONFIG_KEY_0", "protocol.file.allow")
		setEnvForSpec("GIT_CONFIG_VALUE_0", "always")
		var output bytes.Buffer
		gomega.Expect(ArchiveWithSubmodules(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "sub",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{"plain.txt": "ordinary submodule bytes"}))
	})

	ginkgo.DescribeTable("checks every selected pointer size when OIDs repeat",
		func(ctx ginkgo.SpecContext, secondSize int) {
			writeLFSTestFile(repoDir, "a.bin", content, true)
			pointer := strings.Replace(lfsTestPointer(content), fmt.Sprintf("size %d", len(content)), fmt.Sprintf("size %d", secondSize), 1)
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "b.bin"), []byte(pointer), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "repeated OID")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+repoDir)
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit:      gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"),
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})
			if secondSize != len(content) {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("b.bin"))
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("size"))
				return
			}
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{"a.bin": content, "b.bin": content}))
		},
		ginkgo.Entry("matching sizes", len(content)),
		ginkgo.Entry("contradictory sizes", len(content)+1),
	)

	ginkgo.DescribeTable("recognizes committed pointers despite host attribute overrides",
		func(ctx ginkgo.SpecContext, source string) {
			writeLFSTestFile(repoDir, "payload.bin", content, true)
			if source == "info" {
				gomega.Expect(os.WriteFile(filepath.Join(repoDir, ".gitattributes"), []byte("*.bin filter=lfs\n"), 0o644)).To(gomega.Succeed())
			}
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "committed LFS pointer")
			gitSucceed(ctx, repoDir, "remote", "add", "origin", "file://"+repoDir)
			attributes := filepath.Join(repoDir, ".git", "info", "attributes")
			if source == "global" {
				attributes = filepath.Join(SuiteData.TestDirPath, "host-attributes")
				gitSucceed(ctx, repoDir, "config", "core.attributesFile", attributes)
			} else {
				gomega.Expect(os.MkdirAll(filepath.Dir(attributes), 0o755)).To(gomega.Succeed())
			}
			gomega.Expect(os.WriteFile(attributes, []byte("*.bin -filter\n"), 0o644)).To(gomega.Succeed())
			gomega.Expect(gitSucceedTrimmed(ctx, repoDir, "check-attr", "filter", "--", "payload.bin")).To(gomega.Equal("payload.bin: filter: unset"))
			var output bytes.Buffer
			gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin",
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})).To(gomega.Succeed())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": content}))
		},
		ginkgo.Entry("info attributes", "info"),
		ginkgo.Entry("core.attributesFile", "global"),
	)

	ginkgo.DescribeTable("handles unsupported and malformed pointer text through Archive",
		func(ctx ginkgo.SpecContext, pointer string, extension bool) {
			gomega.Expect(os.WriteFile(filepath.Join(repoDir, "payload.bin"), []byte(pointer), 0o644)).To(gomega.Succeed())
			gitSucceed(ctx, repoDir, "add", ".")
			gitCommitSucceed(ctx, repoDir, "-m", "unsupported pointer")
			var output bytes.Buffer
			err := Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
				Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "payload.bin",
				PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
			})
			if extension {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("payload.bin"))
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("extensions are not supported"))
				return
			}
			cmd := exec.CommandContext(ctx, "git", "lfs", "pointer", "--check", "--file=payload.bin")
			cmd.Dir = repoDir
			gomega.Expect(cmd.Run()).To(gomega.BeAssignableToTypeOf(&exec.ExitError{}))
			gomega.Expect(nativeLFSTestSmudge(ctx, repoDir, "payload.bin", pointer)).To(gomega.Equal(pointer))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": pointer}))
		},
		ginkgo.Entry("extension", strings.Replace(lfsTestPointer(content), "oid ", fmt.Sprintf("ext-0-custom sha256:%x\noid ", sha256.Sum256([]byte(content))), 1), true),
		ginkgo.Entry("invalid object id", strings.Replace(lfsTestPointer(content), "sha256:", "sha256:z", 1), false),
		ginkgo.Entry("negative size", strings.Replace(lfsTestPointer(content), fmt.Sprintf("size %d", len(content)), "size -1", 1), false),
	)

	ginkgo.It("exports ordinary text unchanged", func(ctx ginkgo.SpecContext) {
		gomega.Expect(os.WriteFile(filepath.Join(repoDir, "ordinary.txt"), []byte("ordinary text"), 0o644)).To(gomega.Succeed())
		gitSucceed(ctx, repoDir, "add", ".")
		gitCommitSucceed(ctx, repoDir, "-m", "ordinary text")
		var output bytes.Buffer
		gomega.Expect(Archive(ctx, &output, filepath.Join(repoDir, ".git"), cacheDir, ArchiveOptions{
			Commit: gitSucceedTrimmed(ctx, repoDir, "rev-parse", "HEAD"), PathScope: "ordinary.txt",
			PathMatcher: path_matcher.NewPathMatcher(path_matcher.PathMatcherOptions{}),
		})).To(gomega.Succeed())
		gomega.Expect(readTestTar(output.Bytes())).To(gomega.Equal(map[string]string{".": "ordinary text"}))
	})
})

var _ = ginkgo.Describe("LFS credential logging", func() {
	ginkgo.DescribeTable("redacts all credential fields",
		func(format string) {
			credentials := LFSCredentials{URL: "https://url-user:url-secret@example.invalid/repo", Username: "auth-user", Password: "auth-secret"}
			for _, value := range []any{credentials, &credentials, ArchiveOptions{LFSCredentials: &credentials}} {
				formatted := fmt.Sprintf(format, value)
				for _, secret := range []string{"url-user", "url-secret", "auth-user", "auth-secret"} {
					gomega.Expect(formatted).NotTo(gomega.ContainSubstring(secret))
				}
			}
		},
		ginkgo.Entry("default", "%v"),
		ginkgo.Entry("fields", "%+v"),
		ginkgo.Entry("Go syntax", "%#v"),
	)
})
