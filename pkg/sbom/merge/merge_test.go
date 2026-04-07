package merge

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SBOM merge helpers", func() {
	DescribeTable("ReadInputMapping validation",
		func(raw, expectedErr string) {
			tempDir := GinkgoT().TempDir()
			path := filepath.Join(tempDir, "mapping.json")
			Expect(os.WriteFile(path, []byte(raw), 0o644)).To(Succeed())

			mapping, err := ReadInputMapping(path)
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
				Expect(mapping).To(HaveLen(1))
				Expect(mapping).To(HaveKey("frontend"))
				return
			}

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErr))
		},
		Entry("success", `{"frontend":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`, ""),
		Entry("invalid json", `{"frontend":`, "parse JSON"),
		Entry("empty mapping", `{}`, "empty mapping"),
		Entry("invalid digest", `{"frontend":"not-a-digest"}`, "invalid digest"),
		Entry("empty image name", `{"":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`, "image name must not be empty"),
	)

	It("WriteOutput writes file content", func() {
		tempDir := GinkgoT().TempDir()
		outputPath := filepath.Join(tempDir, "result.json")

		Expect(WriteOutput([]byte(`{"ok":true}`), outputPath)).To(Succeed())

		data, err := os.ReadFile(outputPath)
		Expect(err).NotTo(HaveOccurred())

		var payload map[string]bool
		Expect(json.Unmarshal(data, &payload)).To(Succeed())
		Expect(payload).To(HaveKeyWithValue("ok", true))
	})
})
