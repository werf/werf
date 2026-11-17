package instruction

import (
	"context"
	"os"
	"syscall"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/buildah"
	"github.com/werf/werf/v3/pkg/container_backend"
	"github.com/werf/werf/v3/pkg/werf"
)

var _ = ginkgo.Describe("heredoc sources", func() {
	ginkgo.BeforeEach(func() {
		gomega.Expect(werf.Init(ginkgo.GinkgoT().TempDir(), ginkgo.GinkgoT().TempDir())).To(gomega.Succeed())
	})

	ginkgo.DescribeTable("are written to files and copied to the destination",
		func(newInstruction func(instructions.SourcesAndDest, string, string) container_backend.InstructionInterface, sourcesAndDest instructions.SourcesAndDest, expectedCopies []heredocCopyCall, expectedAdds [][]string) {
			backend := &heredocBackendStub{}
			err := newInstruction(sourcesAndDest, "1000:1000", "755").Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(backend.copies).To(gomega.Equal(expectedCopies))
			gomega.Expect(backend.adds).To(gomega.Equal(expectedAdds))
			for _, dir := range backend.heredocDirs {
				gomega.Expect(dir).NotTo(gomega.BeADirectory())
			}
		},
		ginkgo.Entry("COPY with a single heredoc",
			newHeredocTestCopy,
			instructions.SourcesAndDest{DestPath: "/install.sh", SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "echo hi\n"}}},
			[]heredocCopyCall{{Sources: []string{"EOF"}, Dest: "/install.sh", Chown: "1000:1000", Chmod: "755", Contents: map[string]string{"EOF": "echo hi\n"}}},
			nil,
		),
		ginkgo.Entry("COPY with context files and several heredocs",
			newHeredocTestCopy,
			instructions.SourcesAndDest{DestPath: "/dst/", SourcePaths: []string{"src"}, SourceContents: []instructions.SourceContent{{Path: "A", Data: "a\n"}, {Path: "B", Data: "b\n"}}},
			[]heredocCopyCall{
				{Dir: "/source", Sources: []string{"src"}, Dest: "/dst/", Chown: "1000:1000", Chmod: "755"},
				{Sources: []string{"A", "B"}, Dest: "/dst/", Chown: "1000:1000", Chmod: "755", Contents: map[string]string{"A": "a\n", "B": "b\n"}},
			},
			nil,
		),
		ginkgo.Entry("ADD with a single heredoc",
			newHeredocTestAdd,
			instructions.SourcesAndDest{DestPath: "/install.sh", SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "echo hi\n"}}},
			[]heredocCopyCall{{Sources: []string{"EOF"}, Dest: "/install.sh", Chown: "1000:1000", Chmod: "755", Contents: map[string]string{"EOF": "echo hi\n"}}},
			nil,
		),
		ginkgo.Entry("ADD with context files and a heredoc",
			newHeredocTestAdd,
			instructions.SourcesAndDest{DestPath: "/dst/", SourcePaths: []string{"src"}, SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "echo hi\n"}}},
			[]heredocCopyCall{{Sources: []string{"EOF"}, Dest: "/dst/", Chown: "1000:1000", Chmod: "755", Contents: map[string]string{"EOF": "echo hi\n"}}},
			[][]string{{"src"}},
		),
	)

	ginkgo.It("writes heredoc sources with mode 0644 regardless of the umask", func() {
		oldUmask := syscall.Umask(0o077)
		defer syscall.Umask(oldUmask)

		backend := &heredocBackendStub{}
		err := newHeredocTestCopy(instructions.SourcesAndDest{DestPath: "/dst/", SourceContents: []instructions.SourceContent{{Path: "EOF", Data: "x"}}}, "", "").
			Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(backend.heredocModes).To(gomega.Equal(map[string]os.FileMode{"EOF": 0o644}))
	})

	ginkgo.It("rejects a heredoc name escaping the sources dir", func() {
		backend := &heredocBackendStub{}
		err := newHeredocTestCopy(instructions.SourcesAndDest{DestPath: "/dst/", SourceContents: []instructions.SourceContent{{Path: "../EOF", Data: "x"}}}, "", "").
			Apply(context.Background(), "destination", backend, buildah.CommonOpts{}, &copyArchiveStub{})
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(`invalid heredoc name "../EOF"`)))
		gomega.Expect(backend.copies).To(gomega.BeEmpty())
	})
})

func newHeredocTestCopy(sourcesAndDest instructions.SourcesAndDest, chown, chmod string) container_backend.InstructionInterface {
	return NewCopy(instructions.CopyCommand{SourcesAndDest: sourcesAndDest, Chown: chown, Chmod: chmod})
}

func newHeredocTestAdd(sourcesAndDest instructions.SourcesAndDest, chown, chmod string) container_backend.InstructionInterface {
	return NewAdd(instructions.AddCommand{SourcesAndDest: sourcesAndDest, Chown: chown, Chmod: chmod})
}
