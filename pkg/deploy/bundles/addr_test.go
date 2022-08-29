package bundles

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bundle addr", func() {
	It("should parse different bundle address schemas", func() {
		{
			addr, err := ParseAddr("registry.werf.io/group/image:mytag")
			Expect(err).NotTo(HaveOccurred())
			Expect(addr.RegistryAddress).NotTo(BeNil())
			Expect(addr.ArchiveAddress).To(BeNil())
			Expect(addr.Repo).To(Equal("registry.werf.io/group/image"))
			Expect(addr.Tag).To(Equal("mytag"))
		}

		{
			addr, err := ParseAddr("registry.werf.io/group/image")
			Expect(err).NotTo(HaveOccurred())
			Expect(addr.RegistryAddress).NotTo(BeNil())
			Expect(addr.ArchiveAddress).To(BeNil())
			Expect(addr.Repo).To(Equal("registry.werf.io/group/image"))
			Expect(addr.Tag).To(Equal("latest"))
		}

		{
			addr, err := ParseAddr("archive:path/to/file.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			Expect(addr.RegistryAddress).To(BeNil())
			Expect(addr.ArchiveAddress).NotTo(BeNil())
			Expect(addr.Path).To(Equal("path/to/file.tar.gz"))
		}

		{
			addr, err := ParseAddr("archive:/absolute/path/to/file.tar.gz")
			Expect(err).NotTo(HaveOccurred())
			Expect(addr.RegistryAddress).To(BeNil())
			Expect(addr.ArchiveAddress).NotTo(BeNil())
			Expect(addr.Path).To(Equal("/absolute/path/to/file.tar.gz"))
		}
	})
})
