package docker_registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ParseReferencePartsEntry struct {
	reference   string
	expectation referenceParts
}

var _ = DescribeTable("Api_ParseReferenceParts", func(entry ParseReferencePartsEntry) {
	parts, err := (&api{}).parseReferenceParts(entry.reference)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(parts).Should(Equal(entry.expectation))
},
	Entry("account/project", ParseReferencePartsEntry{
		reference: "account/project",
		expectation: referenceParts{
			registry:   name.DefaultRegistry,
			repository: "account/project",
			tag:        name.DefaultTag,
		},
	}),
	Entry("repo", ParseReferencePartsEntry{
		reference: "repo",
		expectation: referenceParts{
			registry:   name.DefaultRegistry,
			repository: "library/repo",
			tag:        name.DefaultTag,
		},
	}),
	Entry("registry.com/repo", ParseReferencePartsEntry{
		reference: "registry.com/repo",
		expectation: referenceParts{
			registry:   "registry.com",
			repository: "repo",
			tag:        name.DefaultTag,
		},
	}),
	Entry("registry.com/repo:tag", ParseReferencePartsEntry{
		reference: "registry.com/repo:tag",
		expectation: referenceParts{
			registry:   "registry.com",
			repository: "repo",
			tag:        "tag",
		},
	}),
	Entry("registry.com/repo:tag@sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209", ParseReferencePartsEntry{
		reference: "registry.com/repo:tag@sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209",
		expectation: referenceParts{
			registry:   "registry.com",
			repository: "repo",
			tag:        "tag",
			digest:     "sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209",
		},
	}),
)

var _ = Describe("api.isInsecureHost", func() {
	var testApi *api

	BeforeEach(func() {
		testApi = newAPI(apiOptions{
			InsecureRegistryHosts: []string{
				"my-registry.local:5000",
				"10.0.0.100",
				"192.168.1.0/24",
			},
		})
	})

	It("should match exact host:port", func() {
		Expect(testApi.isInsecureHost("my-registry.local:5000")).To(BeTrue())
	})

	It("should match IP in CIDR range", func() {
		Expect(testApi.isInsecureHost("192.168.1.50:8080")).To(BeTrue())
	})

	It("should not match host outside list", func() {
		Expect(testApi.isInsecureHost("docker.io")).To(BeFalse())
	})

	It("should return true for all hosts when InsecureRegistry flag is set", func() {
		apiWithFlag := newAPI(apiOptions{InsecureRegistry: true})
		Expect(apiWithFlag.isInsecureHost("any-registry.com")).To(BeTrue())
	})
})
