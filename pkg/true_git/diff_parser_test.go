package true_git

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/path_matcher"
)

var _ = Describe("diffParser", func() {
	var (
		out         *bytes.Buffer
		parser      *diffParser
		pathMatcher path_matcher.PathMatcher
	)

	BeforeEach(func() {
		out = &bytes.Buffer{}
		pathMatcher = path_matcher.NewTruePathMatcher()
	})

	Describe("handleShortBinaryHeader", func() {
		Context("when deleting a binary file with short format (Binary files ... differ)", func() {
			It("should add path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// Simulate parsing a deleted binary file diff
				// diff --git a/old/logo.png b/old/logo.png
				// deleted file mode 100644
				// index abc123..000000
				// Binary files a/old/logo.png and /dev/null differ

				err := parser.handleDiffLine("diff --git a/old/logo.png b/old/logo.png")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(diffBegin))

				err = parser.handleDiffLine("deleted file mode 100644")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(deleteFileDiff))

				err = parser.handleDiffLine("index abc12345..00000000")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(deleteFileDiff))

				err = parser.handleDiffLine("Binary files a/old/logo.png and /dev/null differ")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.Paths).To(ContainElement("old/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("old/logo.png"))
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
			})
		})

		Context("when creating a new binary file with short format", func() {
			It("should NOT add path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// Simulate parsing a new binary file diff
				// diff --git a/new/logo.png b/new/logo.png
				// new file mode 100644
				// index 000000..abc123
				// Binary files /dev/null and b/new/logo.png differ

				err := parser.handleDiffLine("diff --git a/new/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("new file mode 100644")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(newFileDiff))

				err = parser.handleDiffLine("index 00000000..abc12345")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("Binary files /dev/null and b/new/logo.png differ")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.Paths).To(ContainElement("new/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("new/logo.png"))
				Expect(parser.PathsToRemove).To(BeEmpty())
			})
		})

		Context("when modifying a binary file with short format", func() {
			It("should NOT add path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// Simulate parsing a modified binary file diff
				// diff --git a/file.png b/file.png
				// index abc123..def456
				// Binary files a/file.png and b/file.png differ

				err := parser.handleDiffLine("diff --git a/file.png b/file.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("index abc12345..def45678")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(modifyFileDiff))

				err = parser.handleDiffLine("Binary files a/file.png and b/file.png differ")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.Paths).To(ContainElement("file.png"))
				Expect(parser.BinaryPaths).To(ContainElement("file.png"))
				Expect(parser.PathsToRemove).To(BeEmpty())
			})
		})
	})

	Describe("handleBinaryBeginHeader", func() {
		Context("when deleting a binary file with GIT binary patch format", func() {
			It("should add path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// Simulate parsing a deleted binary file diff with --binary
				// diff --git a/old/logo.png b/old/logo.png
				// deleted file mode 100644
				// index abc123..000000
				// GIT binary patch
				// literal 0
				// ...

				err := parser.handleDiffLine("diff --git a/old/logo.png b/old/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("deleted file mode 100644")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(deleteFileDiff))

				err = parser.handleDiffLine("index abc12345..00000000")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("GIT binary patch")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.Paths).To(ContainElement("old/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("old/logo.png"))
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
			})
		})

		Context("when creating a new binary file with GIT binary patch format", func() {
			It("should NOT add path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				err := parser.handleDiffLine("diff --git a/new/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("new file mode 100644")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(newFileDiff))

				err = parser.handleDiffLine("index 00000000..abc12345")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("GIT binary patch")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.Paths).To(ContainElement("new/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("new/logo.png"))
				Expect(parser.PathsToRemove).To(BeEmpty())
			})
		})
	})

	Describe("binary file rename scenario (diff.renames=false)", func() {
		Context("when a binary file is renamed (shown as delete + create)", func() {
			It("should add old path to PathsToRemove and both paths to Paths", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// With diff.renames=false, a rename is shown as two separate diffs:
				// 1. New file creation
				// 2. Old file deletion

				// First: new file creation
				err := parser.handleDiffLine("diff --git a/new/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("new file mode 100644")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("index 00000000..abc12345")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("Binary files /dev/null and b/new/logo.png differ")
				Expect(err).NotTo(HaveOccurred())

				// Second: old file deletion
				err = parser.handleDiffLine("diff --git a/old/logo.png b/old/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("deleted file mode 100644")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("index abc12345..00000000")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("Binary files a/old/logo.png and /dev/null differ")
				Expect(err).NotTo(HaveOccurred())

				// Both paths should be in Paths
				Expect(parser.Paths).To(ContainElement("new/logo.png"))
				Expect(parser.Paths).To(ContainElement("old/logo.png"))

				// Both paths should be in BinaryPaths
				Expect(parser.BinaryPaths).To(ContainElement("new/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("old/logo.png"))

				// Only old path should be in PathsToRemove
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
				Expect(parser.PathsToRemove).NotTo(ContainElement("new/logo.png"))
			})
		})
	})

	Describe("text file deletion", func() {
		Context("when deleting a text file", func() {
			It("should add path to PathsToRemove via handleDeleteFilePath", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// diff --git a/old/file.txt b/old/file.txt
				// deleted file mode 100644
				// index abc123..000000
				// --- a/old/file.txt
				// +++ /dev/null
				// @@ ...

				err := parser.handleDiffLine("diff --git a/old/file.txt b/old/file.txt")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("deleted file mode 100644")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(deleteFileDiff))

				err = parser.handleDiffLine("index abc12345..00000000")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("--- a/old/file.txt")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.PathsToRemove).To(ContainElement("old/file.txt"))
			})
		})
	})

	Describe("pathScope handling", func() {
		Context("when pathScope is set", func() {
			It("should trim pathScope from file paths in PathsToRemove", func() {
				parser = makeDiffParser(out, "assets/logos", pathMatcher, nil)

				err := parser.handleDiffLine("diff --git a/assets/logos/old.png b/assets/logos/old.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("deleted file mode 100644")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("index abc12345..00000000")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("Binary files a/assets/logos/old.png and /dev/null differ")
				Expect(err).NotTo(HaveOccurred())

				// Path should be relative to pathScope
				Expect(parser.Paths).To(ContainElement("old.png"))
				Expect(parser.PathsToRemove).To(ContainElement("old.png"))
			})
		})
	})

	Describe("rename with similarity index format (diff.renames=true)", func() {
		Context("when diff has different paths in a/ and b/ (rename format)", func() {
			It("should add both old and new paths to Paths", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// When git uses rename detection, it shows:
				// diff --git a/old/path b/new/path
				// similarity index 100%
				// rename from old/path
				// rename to new/path

				err := parser.handleDiffLine("diff --git a/old/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(diffBegin))

				// Both paths should be in Paths after handleDiffBegin
				Expect(parser.Paths).To(ContainElement("old/logo.png"))
				Expect(parser.Paths).To(ContainElement("new/logo.png"))

				// similarity index line transitions to renameDiff state
				err = parser.handleDiffLine("similarity index 100%")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(renameDiff))

				// rename from line adds old path to PathsToRemove
				err = parser.handleDiffLine("rename from old/logo.png")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))

				err = parser.handleDiffLine("rename to new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// Old path should be in PathsToRemove for proper removal during patch
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
				Expect(parser.PathsToRemove).NotTo(ContainElement("new/logo.png"))
			})
		})

		Context("when diff has different paths with binary content (rename with modification)", func() {
			It("should add paths correctly", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// This scenario shouldn't happen with diff.renames=false,
				// but if it does, we should handle it gracefully

				err := parser.handleDiffLine("diff --git a/old/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// Both paths should be in Paths
				Expect(parser.Paths).To(ContainElement("old/logo.png"))
				Expect(parser.Paths).To(ContainElement("new/logo.png"))

				// If there's an index line, we go to modifyFileDiff state
				err = parser.handleDiffLine("index abc12345..def67890")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(modifyFileDiff))

				// Binary files line in modifyFileDiff state
				err = parser.handleDiffLine("Binary files a/old/logo.png and b/new/logo.png differ")
				Expect(err).NotTo(HaveOccurred())

				// Both paths should be in BinaryPaths
				Expect(parser.BinaryPaths).To(ContainElement("old/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("new/logo.png"))

				// In modifyFileDiff state, paths should NOT be added to PathsToRemove
				Expect(parser.PathsToRemove).To(BeEmpty())
			})
		})

		Context("when binary file is renamed with 100% similarity (real user scenario)", func() {
			It("should add source path to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// This is the exact scenario from the bug report:
				// diff --git a/assets/logos/units/deckhouse/color_icon_null_null/logo.png b/assets/logos/units/deckhouse/color_icon_light/logo.png
				// similarity index 100%
				// rename from assets/logos/units/deckhouse/color_icon_null_null/logo.png
				// rename to assets/logos/units/deckhouse/color_icon_light/logo.png

				err := parser.handleDiffLine("diff --git a/assets/logos/units/deckhouse/color_icon_null_null/logo.png b/assets/logos/units/deckhouse/color_icon_light/logo.png")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(diffBegin))

				Expect(parser.Paths).To(ContainElement("assets/logos/units/deckhouse/color_icon_null_null/logo.png"))
				Expect(parser.Paths).To(ContainElement("assets/logos/units/deckhouse/color_icon_light/logo.png"))

				err = parser.handleDiffLine("similarity index 100%")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(renameDiff))

				err = parser.handleDiffLine("rename from assets/logos/units/deckhouse/color_icon_null_null/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("rename to assets/logos/units/deckhouse/color_icon_light/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// The old path must be in PathsToRemove so that werf removes it from the image
				Expect(parser.PathsToRemove).To(ContainElement("assets/logos/units/deckhouse/color_icon_null_null/logo.png"))
				Expect(parser.PathsToRemove).NotTo(ContainElement("assets/logos/units/deckhouse/color_icon_light/logo.png"))
			})
		})

		Context("when pathScope is set for rename", func() {
			It("should correctly trim pathScope from rename paths", func() {
				parser = makeDiffParser(out, "assets/logos/units/deckhouse", pathMatcher, nil)

				err := parser.handleDiffLine("diff --git a/assets/logos/units/deckhouse/color_icon_null_null/logo.png b/assets/logos/units/deckhouse/color_icon_light/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("similarity index 100%")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("rename from assets/logos/units/deckhouse/color_icon_null_null/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// Path should be relative to pathScope
				Expect(parser.PathsToRemove).To(ContainElement("color_icon_null_null/logo.png"))
			})
		})

		Context("when multiple files are renamed", func() {
			It("should add all source paths to PathsToRemove", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				// First rename
				err := parser.handleDiffLine("diff --git a/old/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("similarity index 100%")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("rename from old/logo.png")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("rename to new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// Second rename
				err = parser.handleDiffLine("diff --git a/old/icon.svg b/new/icon.svg")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("similarity index 100%")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("rename from old/icon.svg")
				Expect(err).NotTo(HaveOccurred())
				err = parser.handleDiffLine("rename to new/icon.svg")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
				Expect(parser.PathsToRemove).To(ContainElement("old/icon.svg"))
				Expect(parser.PathsToRemove).NotTo(ContainElement("new/logo.png"))
				Expect(parser.PathsToRemove).NotTo(ContainElement("new/icon.svg"))
			})
		})

		Context("when rename has content modification (similarity < 100%)", func() {
			It("should handle rename with index and binary patch", func() {
				parser = makeDiffParser(out, "", pathMatcher, nil)

				err := parser.handleDiffLine("diff --git a/old/logo.png b/new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				err = parser.handleDiffLine("similarity index 95%")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(renameDiff))

				err = parser.handleDiffLine("rename from old/logo.png")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))

				err = parser.handleDiffLine("rename to new/logo.png")
				Expect(err).NotTo(HaveOccurred())

				// After rename lines, index line transitions to modifyFileDiff
				err = parser.handleDiffLine("index abc12345..def67890")
				Expect(err).NotTo(HaveOccurred())
				Expect(parser.state).To(Equal(modifyFileDiff))

				// Then binary patch
				err = parser.handleDiffLine("GIT binary patch")
				Expect(err).NotTo(HaveOccurred())

				Expect(parser.BinaryPaths).To(ContainElement("old/logo.png"))
				Expect(parser.BinaryPaths).To(ContainElement("new/logo.png"))
				Expect(parser.PathsToRemove).To(ContainElement("old/logo.png"))
			})
		})
	})
})
