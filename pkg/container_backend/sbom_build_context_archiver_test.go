package container_backend_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/container_backend"
)

var _ = Describe("SbomBuildContextArchiver", func() {
	DescribeTable("Create() should",
		func(ctx SpecContext, tc testCase) {
			tmpDir := GinkgoT().TempDir()

			// arrange: create input files
			prepareFixtures(tmpDir, tc.fixtures)

			archiver := container_backend.NewSbomContextArchiver(tmpDir)

			err := archiver.Create(ctx, container_backend.BuildContextArchiveCreateOptions{
				ContextAddFiles: tc.contextAdd,
			})
			Expect(err).To(tc.createErrMatcher)

			Expect(archiver.Path()).To(Equal(filepath.Join(tmpDir, "sbom-docker.tar")))

			verifyArchive(tmpDir, tc)
		},
		Entry("successfully creates tar and adds files from mode 0600", testCase{
			fixtures: map[string][]byte{
				"foo.txt":       []byte("hello"),
				"dir/bar.json":  []byte(`{"x":1}`),
				"empty.bin":     []byte(""),
				"dir/nested.md": []byte("# t"),
			},
			contextAdd: []string{
				"foo.txt",
				"dir/bar.json",
				"empty.bin",
			},
			createErrMatcher: Succeed(),
			verifyArchive:    true,
			verifyEntries: map[string][]byte{
				"foo.txt":      []byte("hello"),
				"dir/bar.json": []byte(`{"x":1}`),
				"empty.bin":    []byte(""),
			},
		}),
		Entry("error if the added file is missing", testCase{
			fixtures: map[string][]byte{
				"exists.txt": []byte("ok"),
			},
			contextAdd: []string{
				"exists.txt",
				"nope.txt",
			},
			createErrMatcher: MatchError(ContainSubstring("nope.txt: no such file or directory")),
			verifyArchive:    false,
		}),
	)
})

type testCase struct {
	fixtures         map[string][]byte // path -> content (relative to root)
	contextAdd       []string
	createErrMatcher types.GomegaMatcher
	verifyArchive    bool
	verifyEntries    map[string][]byte // name -> content
}

func prepareFixtures(tmpDir string, fixtures map[string][]byte) {
	// arrange: create input files
	for rel, content := range fixtures {
		abs := filepath.Join(tmpDir, rel)
		Expect(os.MkdirAll(filepath.Dir(abs), 0o755)).To(Succeed())
		Expect(os.WriteFile(abs, content, 0o644)).To(Succeed())
	}
}

func verifyArchive(tmpDir string, tc testCase) {
	if tc.verifyArchive {
		f, openErr := os.Open(filepath.Join(tmpDir, "sbom-docker.tar"))
		Expect(openErr).NotTo(HaveOccurred())
		defer f.Close()

		tr := tar.NewReader(f)
		seen := map[string]struct{}{}

		for {
			hdr, rErr := tr.Next()
			if errors.Is(rErr, io.EOF) {
				break
			}
			Expect(rErr).NotTo(HaveOccurred())

			seen[hdr.Name] = struct{}{}

			Expect(hdr.Mode).To(Equal(int64(0o600)))

			if want, ok := tc.verifyEntries[hdr.Name]; ok {
				got, readErr := io.ReadAll(tr)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(bytes.Equal(got, want)).To(BeTrue(), "content mismatch for %s", hdr.Name)
				Expect(hdr.Size).To(Equal(int64(len(want))))
			}
		}

		for name := range tc.verifyEntries {
			_, ok := seen[name]
			Expect(ok).To(BeTrue(), "expected tar entry %q to exist", name)
		}
	}
}
