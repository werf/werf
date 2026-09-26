package docker_registry

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TryLoadTagsFromCacheEntry struct {
	age      time.Duration
	opts     []Option
	expectOk bool
}

var _ = DescribeTable("tryLoadTagsFromCache", func(entry TryLoadTagsFromCacheEntry) {
	r := &DockerRegistryWithCache{cachedTagsMap: &sync.Map{}}
	r.cachedTagsMap.Store("repo", tagsCacheEntry{tags: []string{"tag"}, updatedAt: time.Now().Add(-entry.age)})

	tags, ok := r.tryLoadTagsFromCache("repo", entry.opts...)

	Expect(ok).Should(Equal(entry.expectOk))
	if entry.expectOk {
		Expect(tags).Should(Equal([]string{"tag"}))
	} else {
		Expect(tags).Should(BeNil())
	}
},
	Entry("no options: fresh entry is not served", TryLoadTagsFromCacheEntry{
		age:      time.Millisecond,
		expectOk: false,
	}),
	Entry("max age: entry younger than max age is served", TryLoadTagsFromCacheEntry{
		age:      time.Millisecond,
		opts:     []Option{WithTagsMaxAge(time.Minute)},
		expectOk: true,
	}),
	Entry("max age: entry older than max age is not served", TryLoadTagsFromCacheEntry{
		age:      time.Minute,
		opts:     []Option{WithTagsMaxAge(time.Millisecond)},
		expectOk: false,
	}),
	Entry("cached tags: stale entry is served regardless of age", TryLoadTagsFromCacheEntry{
		age:      time.Hour,
		opts:     []Option{WithCachedTags(), WithTagsMaxAge(time.Millisecond)},
		expectOk: true,
	}),
)

var _ = Describe("tryLoadTagsFromCache", func() {
	It("should not serve a missing entry", func() {
		r := &DockerRegistryWithCache{cachedTagsMap: &sync.Map{}}

		tags, ok := r.tryLoadTagsFromCache("repo", WithTagsMaxAge(time.Minute))

		Expect(ok).Should(BeFalse())
		Expect(tags).Should(BeNil())
	})
})
