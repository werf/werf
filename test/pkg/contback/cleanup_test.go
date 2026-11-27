package contback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestCleanup(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.DescribeTable("Cleanup retry ceiling", func(progress bool) {
		synctest.Test(t, func(_ *testing.T) {
			defer ginkgo.GinkgoRecover()
			ctx, cancel := context.WithTimeout(context.Background(), 110*time.Second)
			defer cancel()
			attempts := 0
			started := time.Now()
			inUse := fmt.Errorf("container owner-id: %w", errCleanupImageInUse)
			err := cleanupImageReferences(ctx, "buildah", io.Discard, func() ([]string, error) {
				if progress {
					if attempts > 101 {
						return nil, errors.New("cleanup exceeded pass ceiling")
					}
					return []string{fmt.Sprintf("owned-%d", attempts)}, nil
				}
				return []string{"owned"}, nil
			}, func(context.Context, string) error {
				attempts++
				if !progress {
					<-time.After(2 * time.Second)
				}
				return inUse
			})
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("test images remain")))
			gomega.Expect(errors.Is(err, context.DeadlineExceeded)).To(gomega.BeFalse())
			if progress {
				gomega.Expect(attempts).To(gomega.Equal(101))
				gomega.Expect(time.Since(started)).To(gomega.BeZero())
			} else {
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("owner-id"))
				gomega.Expect(attempts).To(gomega.Equal(11))
				gomega.Expect(time.Since(started)).To(gomega.Equal(32 * time.Second))
			}
		})
	}, ginkgo.Entry("stalled image-in-use removals", false), ginkgo.Entry("changing inventories", true))
	ginkgo.It("removes a busy image released near the end of the grace period", func() {
		synctest.Test(t, func(_ *testing.T) {
			defer ginkgo.GinkgoRecover()
			var firstConflict time.Time
			refs := []string{"owned"}
			err := cleanupImageReferences(context.Background(), "buildah", io.Discard, func() ([]string, error) { return refs, nil }, func(context.Context, string) error {
				if firstConflict.IsZero() {
					<-time.After(40 * time.Second)
					firstConflict = time.Now()
				}
				if time.Since(firstConflict) < 29*time.Second {
					return errCleanupImageInUse
				}
				refs = nil
				return nil
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(refs).To(gomega.BeEmpty())
		})
	})
	ginkgo.It("starts a new busy window after a reference is removed", func() {
		synctest.Test(t, func(_ *testing.T) {
			defer ginkgo.GinkgoRecover()
			started := time.Now()
			refs := []string{"first", "second"}
			err := cleanupImageReferences(context.Background(), "buildah", io.Discard, func() ([]string, error) { return refs, nil }, func(_ context.Context, ref string) error {
				elapsed := time.Since(started)
				if ref == "first" && elapsed >= 20*time.Second || ref == "second" && elapsed >= 45*time.Second {
					refs = refs[1:]
					return nil
				}
				return errCleanupImageInUse
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(time.Since(started)).To(gomega.Equal(45 * time.Second))
		})
	})
	ginkgo.RunSpecs(t, "Test resource cleanup")
}

var _ = ginkgo.Describe("Project image cleanup", func() {
	ginkgo.It("does not retry an unrelated removal error alongside a busy image", func() {
		attempts := 0
		permanentErr := errors.New("permission denied")
		err := cleanupImageReferences(context.Background(), "buildah", io.Discard, func() ([]string, error) {
			return []string{"busy", "inaccessible"}, nil
		}, func(_ context.Context, ref string) error {
			attempts++
			if ref == "busy" {
				return errCleanupImageInUse
			}
			return permanentErr
		})
		gomega.Expect(errors.Is(err, permanentErr)).To(gomega.BeTrue())
		gomega.Expect(attempts).To(gomega.Equal(2))
	})
	ginkgo.It("does not retry an unchanged no-op reference alongside a busy image", func() {
		attempts := 0
		err := cleanupImageReferences(context.Background(), "buildah", io.Discard, func() ([]string, error) {
			return []string{"busy", "unchanged"}, nil
		}, func(_ context.Context, ref string) error {
			attempts++
			gomega.Expect(attempts).To(gomega.BeNumerically("<=", 2))
			if ref == "busy" {
				return errCleanupImageInUse
			}
			return nil
		})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("test images remain: busy, unchanged")))
	})
	ginkgo.It("waits for a concurrent build to release an image without force", func() {
		remaining := []string{"owned"}
		attempts := 0
		var diagnostics strings.Builder
		err := cleanupImageReferences(context.Background(), "buildah", &diagnostics, func() ([]string, error) {
			return remaining, nil
		}, func(context.Context, string) error {
			attempts++
			if attempts > 1 {
				gomega.Expect(diagnostics.String()).To(gomega.ContainSubstring(`"owned"`))
				gomega.Expect(strings.Count(diagnostics.String(), "container owner-id")).To(gomega.Equal(1))
			}
			if attempts <= 2 {
				return fmt.Errorf("container owner-id: %w", errCleanupImageInUse)
			}
			remaining = nil
			return nil
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remaining).To(gomega.BeEmpty())
		gomega.Expect(attempts).To(gomega.Equal(3))
	})
	ginkgo.It("stops waiting on cancellation and reports the retained image", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := cleanupImageReferences(ctx, "buildah", io.Discard, func() ([]string, error) {
			return []string{"owned"}, nil
		}, func(context.Context, string) error {
			cancel()
			return errCleanupImageInUse
		})
		gomega.Expect(errors.Is(err, context.Canceled)).To(gomega.BeTrue())
		gomega.Expect(errors.Is(err, errCleanupImageInUse)).To(gomega.BeTrue())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("owned"))
	})
	ginkgo.It("does not retry a no-op removal on the basis of an earlier conflict", func() {
		attempts := 0
		err := cleanupImageReferences(context.Background(), "buildah", io.Discard, func() ([]string, error) {
			return []string{"owned"}, nil
		}, func(context.Context, string) error {
			attempts++
			if attempts == 1 {
				return errCleanupImageInUse
			}
			gomega.Expect(attempts).To(gomega.Equal(2))
			return nil
		})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("test images remain: owned")))
	})
	const deletedID = "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	const unrequestedID = "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	ginkgo.DescribeTable("verifies stage cleanup with a final full inventory", func(ctx ginkgo.SpecContext, exitCode int, inspectError, extraID, fallbackReason string) {
		dir := ginkgo.GinkgoT().TempDir()
		script, err := os.ReadFile("testdata/cleanup-backend.sh")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		binary := filepath.Join(dir, "docker")
		gomega.Expect(os.WriteFile(binary, script, 0o700)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		const stageID = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		const otherID = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		const digestID = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		row := func(id, repo, tag string) string {
			return fmt.Sprintf("{\"ID\":%q,\"Repository\":%q,\"Tag\":%q}\n", id, repo, tag)
		}
		foreign := row(otherID, "neighbor", "keep") + fmt.Sprintf("{\"ID\":%q,\"Repository\":\"neighbor\",\"Tag\":\"<none>\",\"Digest\":%q}\n", digestID, digestID)
		inventory := binary + ".inventory"
		ginkgo.GinkgoT().Setenv("CLEANUP_INVENTORY", inventory)
		const project = "werf-test-one"
		remaining := map[string]string{project + ":shared": row(otherID, project, "shared"), project + ":digest": row(digestID, project, "digest"), project + ":stage": row(stageID, project, "stage")}
		wantIDs := []string{stageID, otherID, digestID}
		wantRemoved := []string{project + ":shared", project + ":digest", project + ":stage", stageID}
		if inspectError != "" {
			remaining[project+":deleted"] = row(deletedID, project, "deleted")
			wantIDs = append(wantIDs, deletedID)
			wantRemoved = append(wantRemoved, project+":deleted")
		}
		writeInventory := func() error {
			contents := foreign
			for _, value := range remaining {
				contents += value
			}
			return os.WriteFile(inventory, []byte(contents), 0o600)
		}
		gomega.Expect(writeInventory()).To(gomega.Succeed())
		inspected := []map[string]any{
			{"Id": stageID, "RepoTags": []string{}, "RepoDigests": []string{}},
			{"Id": otherID, "RepoTags": []string{"neighbor:keep"}},
			{"Id": digestID, "RepoDigests": []string{"neighbor@" + digestID}},
		}
		if extraID != "" {
			inspected = append(inspected, map[string]any{"Id": extraID})
		}
		output, err := json.Marshal(inspected)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.WriteFile(binary+".inspect", output, 0o600)).To(gomega.Succeed())
		gomega.Expect(os.WriteFile(binary+".error", []byte(inspectError), 0o600)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("CLEANUP_INSPECT_EXIT", fmt.Sprint(exitCode))
		var removed []string
		remove := func(_ context.Context, ref string) error {
			removed = append(removed, ref)
			delete(remaining, ref)
			if ref == "werf-test-one:stage" {
				remaining[stageID] = row(stageID, "<none>", "<none>")
			} else if ref == stageID {
				gomega.Expect(os.WriteFile(binary+".removed-id", nil, 0o600)).To(gomega.Succeed())
			}
			return writeInventory()
		}
		var diagnostics strings.Builder
		gomega.Expect(cleanupDockerProjectImages(ctx, "werf-test-one", nil, &diagnostics, remove)).To(gomega.Succeed())
		gomega.Expect(removed).To(gomega.ConsistOf(wantRemoved))
		ids, err := os.ReadFile(binary + ".ids")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(strings.Fields(string(ids))).To(gomega.ConsistOf(wantIDs))
		calls, err := os.ReadFile(binary + ".listings")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		log := diagnostics.String()
		if fallbackReason != "" {
			gomega.Expect(string(calls)).To(gomega.Equal("list\nlist\nfinal\n"))
			gomega.Expect(log).To(gomega.ContainSubstring("docker cleanup for werf-test-one: falling back to full inventory"))
			gomega.Expect(log).To(gomega.ContainSubstring(fallbackReason))
		} else {
			gomega.Expect(string(calls)).To(gomega.Equal("list\nfinal\n"))
			gomega.Expect(log).NotTo(gomega.ContainSubstring("falling back to full inventory"))
		}
	}, ginkgo.Entry("two full inventories after successful inspection", 0, "", "", ""),
		ginkgo.Entry("two full inventories when an inspected ID is already gone", 1, "Error response from daemon: No such image: "+deletedID, "", ""),
		ginkgo.Entry("rescans on an unexpected exit code despite a known missing ID", 2, "Error response from daemon: No such image: "+deletedID, "", "exit status 2"),
		ginkgo.Entry("rescans when the missing ID was not requested", 1, "Error response from daemon: No such image: "+unrequestedID, "", unrequestedID),
		ginkgo.Entry("rescans when inspect returns an unrequested ID", 0, "", unrequestedID, unrequestedID),
		ginkgo.Entry("rescans on an unexpected inspect failure", 1, "daemon error", "", "daemon error"))
	ginkgo.It("rescans after deleting a dependent child image", func() {
		dir := ginkgo.GinkgoT().TempDir()
		script, err := os.ReadFile("testdata/cleanup-backend.sh")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), script, 0o700)).To(gomega.Succeed())
		inventoryPath := filepath.Join(dir, "inventory")
		parent := "{\"ID\":\"parent\",\"Repository\":\"werf-test-one\",\"Tag\":\"a\"}\n"
		child := "{\"ID\":\"child\",\"Repository\":\"werf-test-one\",\"Tag\":\"z\"}\n"
		gomega.Expect(os.WriteFile(inventoryPath, []byte(parent+child), 0o600)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		ginkgo.GinkgoT().Setenv("CLEANUP_INVENTORY", inventoryPath)
		childRemoved := false
		removeImage := func(_ context.Context, ref string) error {
			if ref == "werf-test-one:z" {
				childRemoved = true
				return os.WriteFile(inventoryPath, []byte(parent), 0o600)
			}
			if !childRemoved {
				return errors.New("parent has a child")
			}
			_, err := os.Stat(filepath.Join(dir, "docker.ids"))
			gomega.Expect(err).To(gomega.MatchError(os.ErrNotExist))
			return os.WriteFile(inventoryPath, nil, 0o600)
		}
		gomega.Expect(cleanupDockerProjectImages(context.Background(), "werf-test-one", nil, io.Discard, removeImage)).To(gomega.Succeed())
		remaining, err := os.ReadFile(inventoryPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remaining).To(gomega.BeEmpty())
	})
	ginkgo.DescribeTable("executes cleanup and verifies the resulting inventory",
		func(inventory, remaining, expectedRef, inspectFailure string) {
			dir := ginkgo.GinkgoT().TempDir()
			cat, err := exec.LookPath("cat")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(os.Symlink(cat, filepath.Join(dir, "cat"))).To(gomega.Succeed())
			script, err := os.ReadFile("testdata/cleanup-backend.sh")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), script, 0o700)).To(gomega.Succeed())
			inventoryPath := filepath.Join(dir, "inventory")
			callsPath := filepath.Join(dir, "calls")
			gomega.Expect(os.WriteFile(inventoryPath, []byte(inventory), 0o600)).To(gomega.Succeed())
			ginkgo.GinkgoT().Setenv("PATH", dir)
			ginkgo.GinkgoT().Setenv("CLEANUP_INVENTORY", inventoryPath)
			ginkgo.GinkgoT().Setenv("CLEANUP_CALLS", callsPath)
			ginkgo.GinkgoT().Setenv("CLEANUP_REMAINING", remaining)
			ginkgo.GinkgoT().Setenv("WERF_REPO", "")
			ginkgo.GinkgoT().Setenv("WERF_FINAL_REPO", "")
			ginkgo.GinkgoT().Setenv("CLEANUP_FAIL", "1")
			gomega.Expect(CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{})).To(gomega.MatchError(gomega.ContainSubstring("exit status 42")))
			ginkgo.GinkgoT().Setenv("CLEANUP_FAIL", "0")
			ginkgo.GinkgoT().Setenv("CLEANUP_NOOP", "1")
			gomega.Expect(CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{})).To(gomega.MatchError(gomega.ContainSubstring("test images remain")))
			ginkgo.GinkgoT().Setenv("CLEANUP_NOOP", "0")
			gomega.Expect(os.WriteFile(callsPath, nil, 0o600)).To(gomega.Succeed())
			if inspectFailure != "" {
				ginkgo.GinkgoT().Setenv("CLEANUP_INSPECT_EXIT", "2")
				gomega.Expect(os.WriteFile(filepath.Join(dir, "docker.error"), []byte(inspectFailure), 0o600)).To(gomega.Succeed())
			}
			gomega.Expect(CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{})).To(gomega.Succeed())
			if inspectFailure != "" {
				gomega.Expect(ginkgo.CurrentSpecReport().CapturedGinkgoWriterOutput).To(gomega.And(
					gomega.ContainSubstring("docker cleanup for werf-test-one: falling back to full inventory"),
					gomega.ContainSubstring(inspectFailure),
				))
			}
			calls, err := os.ReadFile(callsPath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(calls)).To(gomega.Equal(expectedRef + "\n"))
			actual, err := os.ReadFile(inventoryPath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(actual)).To(gomega.Equal(remaining))
		},
		ginkgo.Entry("Docker preserves a foreign alias", "{\"ID\":\"owned\",\"Repository\":\"localhost:5000/werf-test-one\",\"Tag\":\"stage\"}\n{\"ID\":\"owned\",\"Repository\":\"foreign\",\"Tag\":\"keep\"}\n", "{\"ID\":\"owned\",\"Repository\":\"foreign\",\"Tag\":\"keep\"}\n", "localhost:5000/werf-test-one:stage", ""),
		ginkgo.Entry("reports the inspect fallback reason in the spec log", `{"ID":"owned","Repository":"werf-test-one","Tag":"stage"}`, "", "werf-test-one:stage", "inspect fixture unavailable"),
	)
	ginkgo.DescribeTable("selects only project references",
		func(images []cleanupImage, expected []string) {
			gomega.Expect(projectImageReferences(images, "werf-test-one", []string{"registry.example/custom"})).To(gomega.Equal(expected))
		},
		ginkgo.Entry("local and registry tags", []cleanupImage{{Names: []string{"werf-test-one:stage", "localhost:5000/werf-test-one:stage", "localhost/werf-test-one:other"}}}, []string{"localhost/werf-test-one:other", "localhost:5000/werf-test-one:stage", "werf-test-one:stage"}),
		ginkgo.Entry("foreign aliases and prefix collisions", []cleanupImage{{Names: []string{"werf-test-one:stage", "werf-test-one-other:stage", "other:keep"}}}, []string{"werf-test-one:stage"}),
		ginkgo.Entry("final and custom repo", []cleanupImage{{Names: []string{"registry.example/werf-test-one-final:tag", "registry.example/custom:tag"}}}, []string{"registry.example/custom:tag", "registry.example/werf-test-one-final:tag"}),
		ginkgo.Entry("tagless project image", []cleanupImage{{ID: "sha256:owned"}}, []string{"sha256:owned"}),
		ginkgo.Entry("a tagless row with a foreign alias elsewhere", []cleanupImage{{ID: "owned"}, {ID: "owned", Names: []string{"foreign:keep"}}}, []string(nil)),
		ginkgo.Entry("duplicate tags", []cleanupImage{{Names: []string{"werf-test-one:stage", "werf-test-one:stage"}}}, []string{"werf-test-one:stage"}),
	)
	ginkgo.DescribeTable("normalizes registered Docker Hub repositories", func(repo string) {
		images := []cleanupImage{{Names: []string{"foo/bar:stage", "docker.io/foo/bar:other", "foo/foreign:keep"}}}
		gomega.Expect(projectImageReferences(images, "werf-test-one", []string{repo})).To(gomega.Equal([]string{"docker.io/foo/bar:other", "foo/bar:stage"}))
	},
		ginkgo.Entry("canonical", "docker.io/foo/bar"),
		ginkgo.Entry("familiar", "foo/bar"),
	)
	ginkgo.It("decodes Docker inventories", func() {
		images, err := parseDockerCleanupImages([]byte("{\"ID\":\"sha256:owned\",\"Repository\":\"localhost:5000/werf-test-one\",\"Tag\":\"stage\"}\n"))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(images).To(gomega.Equal([]cleanupImage{{ID: "sha256:owned", Names: []string{"localhost:5000/werf-test-one:stage"}}}))
	})
	ginkgo.It("rejects malformed inventories", func() {
		_, err := parseDockerCleanupImages([]byte("not-json"))
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
	ginkgo.It("preserves foreign digest-only references", func() {
		images, err := parseDockerCleanupImages([]byte(`{"ID":"owned","Repository":"foreign","Tag":"<none>","Digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(projectImageReferences(images, "werf-test-one", nil)).To(gomega.BeEmpty())
	})
	ginkgo.It("refuses an empty or non-test project before invoking a backend", func() {
		for _, project := range []string{"", "production", "werf-test-*"} {
			gomega.Expect(CleanupProject(context.Background(), project, CleanupProjectOptions{})).To(gomega.MatchError(gomega.ContainSubstring("refuse cleanup")))
		}
	})
	ginkgo.It("allows hosts without container CLIs", func() {
		ginkgo.GinkgoT().Setenv("PATH", ginkgo.GinkgoT().TempDir())
		gomega.Expect(CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{})).To(gomega.Succeed())
	})
	ginkgo.It("reports a failing installed Docker CLI", func() {
		dir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), []byte("#!/bin/sh\nexit 42\n"), 0o700)).To(gomega.Succeed())
		ginkgo.GinkgoT().Setenv("PATH", dir)
		gomega.Expect(CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{})).To(gomega.MatchError(gomega.ContainSubstring("exit status 42")))
	})
	ginkgo.DescribeTable("selects the project's backends", func(modes []string, wantDocker, wantBuildah bool) {
		dir := ginkgo.GinkgoT().TempDir()
		gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), []byte("#!/bin/sh\nexit 42\n"), 0o700)).To(gomega.Succeed())
		canStubUnshare := runtime.GOOS == "linux" && os.Geteuid() != 0
		if canStubUnshare {
			gomega.Expect(os.WriteFile(filepath.Join(dir, "buildah"), []byte("#!/bin/sh\ntest \"$1\" = unshare || exit 1\nexit 43\n"), 0o700)).To(gomega.Succeed())
		}
		ginkgo.GinkgoT().Setenv("PATH", dir)
		err := CleanupProject(context.Background(), "werf-test-one", CleanupProjectOptions{BackendModes: modes})
		if !wantDocker && !(wantBuildah && canStubUnshare) {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			return
		}
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exit status 42")).To(gomega.Equal(wantDocker))
		gomega.Expect(strings.Contains(err.Error(), "exit status 43")).To(gomega.Equal(wantBuildah && canStubUnshare))
	}, ginkgo.Entry("default Docker", []string{""}, true, false), ginkgo.Entry("explicit Docker", []string{"docker"}, true, false),
		ginkgo.Entry("native rootless", []string{"native-rootless"}, false, true), ginkgo.Entry("native chroot", []string{"native-chroot"}, false, true),
		ginkgo.Entry("auto", []string{"auto"}, false, true), ginkgo.Entry("default", []string{"default"}, false, true),
		ginkgo.Entry("both modes used before teardown", []string{"docker", "native-rootless"}, true, true),
		ginkgo.Entry("no mode recorded", nil, true, true), ginkgo.Entry("unknown mode", []string{"unknown"}, true, true))
})
