package docker_registry

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

type IsReferencedByTagEntry struct {
	name        string
	tag         string
	expectation bool
}

var _ = DescribeTable("isReferencedByTag", func(entry IsReferencedByTagEntry) {
	info := &image.Info{Name: entry.name, Tag: entry.tag}
	Expect(isReferencedByTag(info)).To(Equal(entry.expectation))
},
	Entry("explicit tag", IsReferencedByTagEntry{
		name:        "registry.example.com/repo:abc123-1773320914455",
		tag:         "abc123-1773320914455",
		expectation: true,
	}),
	Entry("digest-only reference", IsReferencedByTagEntry{
		name:        "registry.example.com/repo@sha256:e674a824ddaee080167ac0789095402209ff4b15e743e0bd344c0e39833420c7",
		tag:         "latest",
		expectation: false,
	}),
	Entry("bare repository with implicit latest", IsReferencedByTagEntry{
		name:        "registry.example.com/repo",
		tag:         "latest",
		expectation: false,
	}),
	Entry("empty tag", IsReferencedByTagEntry{
		name:        "registry.example.com/repo:sometag",
		tag:         "",
		expectation: false,
	}),
	Entry("tag with digest (both present)", IsReferencedByTagEntry{
		name:        "registry.example.com/repo:tag@sha256:e674a824ddaee080167ac0789095402209ff4b15e743e0bd344c0e39833420c7",
		tag:         "tag",
		expectation: false,
	}),
)
