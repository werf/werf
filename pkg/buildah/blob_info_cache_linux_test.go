//go:build cgo

package buildah

import (
	"reflect"

	"github.com/containers/image/v5/pkg/blobinfocache"
	"github.com/containers/image/v5/pkg/blobinfocache/memory"
	imgtypes "github.com/containers/image/v5/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
)

var _ = Describe("blob info cache", func() {
	It("persists digest mappings across cache instances", func() {
		sysCtx := &imgtypes.SystemContext{BlobInfoCacheDir: GinkgoT().TempDir()}
		compressed := digest.FromString("compressed")
		uncompressed := digest.FromString("uncompressed")

		first := blobinfocache.DefaultCache(sysCtx)
		Expect(reflect.TypeOf(first)).NotTo(Equal(reflect.TypeOf(memory.New())), "fell back to a memory-only cache")
		first.RecordDigestUncompressedPair(compressed, uncompressed)

		second := blobinfocache.DefaultCache(sysCtx)
		Expect(second.UncompressedDigest(compressed)).To(Equal(uncompressed))
	})
})
