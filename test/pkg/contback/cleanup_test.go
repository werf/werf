package contback

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestCleanup(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Test resource cleanup")
}

var _ = ginkgo.Describe("Project image cleanup", func() {
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
			return os.WriteFile(inventoryPath, nil, 0o600)
		}
		gomega.Expect(cleanupDockerProjectImages(context.Background(), "werf-test-one", nil, removeImage)).To(gomega.Succeed())
		remaining, err := os.ReadFile(inventoryPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(remaining).To(gomega.BeEmpty())
	})
	ginkgo.DescribeTable("executes cleanup and verifies the resulting inventory",
		func(inventory, remaining, expectedRef string) {
			dir := ginkgo.GinkgoT().TempDir()
			script, err := os.ReadFile("testdata/cleanup-backend.sh")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(os.WriteFile(filepath.Join(dir, "docker"), script, 0o700)).To(gomega.Succeed())
			inventoryPath := filepath.Join(dir, "inventory")
			callsPath := filepath.Join(dir, "calls")
			gomega.Expect(os.WriteFile(inventoryPath, []byte(inventory), 0o600)).To(gomega.Succeed())
			ginkgo.GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			ginkgo.GinkgoT().Setenv("CLEANUP_INVENTORY", inventoryPath)
			ginkgo.GinkgoT().Setenv("CLEANUP_CALLS", callsPath)
			ginkgo.GinkgoT().Setenv("CLEANUP_REMAINING", remaining)
			ginkgo.GinkgoT().Setenv("WERF_REPO", "")
			ginkgo.GinkgoT().Setenv("WERF_FINAL_REPO", "")
			removeImage := func(ctx context.Context, ref string) error {
				_, err := cleanupCommand(ctx, "docker", []string{"rmi", ref})
				return err
			}

			ginkgo.GinkgoT().Setenv("CLEANUP_FAIL", "1")
			gomega.Expect(cleanupDockerProjectImages(context.Background(), "werf-test-one", nil, removeImage)).To(gomega.MatchError(gomega.ContainSubstring("exit status 42")))
			ginkgo.GinkgoT().Setenv("CLEANUP_FAIL", "0")
			ginkgo.GinkgoT().Setenv("CLEANUP_NOOP", "1")
			gomega.Expect(cleanupDockerProjectImages(context.Background(), "werf-test-one", nil, removeImage)).To(gomega.MatchError(gomega.ContainSubstring("test images remain")))
			ginkgo.GinkgoT().Setenv("CLEANUP_NOOP", "0")
			gomega.Expect(os.WriteFile(callsPath, nil, 0o600)).To(gomega.Succeed())
			gomega.Expect(cleanupDockerProjectImages(context.Background(), "werf-test-one", nil, removeImage)).To(gomega.Succeed())
			calls, err := os.ReadFile(callsPath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(calls)).To(gomega.Equal(expectedRef + "\n"))
			actual, err := os.ReadFile(inventoryPath)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(string(actual)).To(gomega.Equal(remaining))
		},
		ginkgo.Entry("Docker preserves a foreign alias", "{\"ID\":\"owned\",\"Repository\":\"localhost:5000/werf-test-one\",\"Tag\":\"stage\"}\n{\"ID\":\"owned\",\"Repository\":\"foreign\",\"Tag\":\"keep\"}\n", "{\"ID\":\"owned\",\"Repository\":\"foreign\",\"Tag\":\"keep\"}\n", "localhost:5000/werf-test-one:stage"),
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
})
