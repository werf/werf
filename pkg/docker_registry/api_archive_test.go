package docker_registry

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("api.PullImageArchive", func(child manifestListChild, expectedErr string) {
	ctx := context.Background()
	fixture := newWritableBearerRegistryFixture()
	DeferCleanup(fixture.server.Close)
	registry := newAPI(apiOptions{InsecureRegistry: true})
	ref := strings.TrimPrefix(fixture.server.URL, "http://") + "/repo:tag"

	parsedRef, err := name.ParseReference(ref, registry.parseReferenceOptionsForHost(ref)...)
	Expect(err).NotTo(HaveOccurred())
	Expect(registry.writeToRemote(ctx, parsedRef, buildChildImage(child))).To(Succeed())

	archive := &bytes.Buffer{}
	err = registry.PullImageArchive(ctx, archive, ref)

	if expectedErr != "" {
		Expect(err).To(MatchError(ContainSubstring(expectedErr)))
		Expect(archive.Len()).To(BeZero(), "a rejected image must not write a partial archive")
		return
	}

	Expect(err).NotTo(HaveOccurred())
	restored, err := tarball.Image(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(archive.Bytes())), nil
	}, nil)
	Expect(err).NotTo(HaveOccurred())
	layers, err := restored.Layers()
	Expect(err).NotTo(HaveOccurred())
	Expect(layers).To(HaveLen(1))
},
	Entry("writes an archive for gzip docker layers", dockerChild, ""),
	Entry("writes an archive for gzip oci layers", ociGzipChild, ""),
	Entry("rejects zstd layers", ociZstdChild, "zstd-compressed layers are not supported"),
)
