package docker_registry

import (
	"errors"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("IsTagNotFoundError", func(err error, expected bool) {
	gomega.Expect(IsTagNotFoundError(err, "marker")).To(gomega.Equal(expected))
},
	ginkgo.Entry("the tag is missing", registryErrorAt(transport.ManifestUnknownErrorCode, "/v2/project/stages/manifests/marker"), true),
	ginkgo.Entry("the repository is missing", registryErrorAt(transport.NameUnknownErrorCode, "/v2/project/stages/manifests/marker"), true),
	ginkgo.Entry("a child manifest of the found tag is missing", registryErrorAt(transport.ManifestUnknownErrorCode, "/v2/project/stages/manifests/sha256:aaaa"), false),
	ginkgo.Entry("a blob of the found tag is missing", registryErrorAt(transport.BlobUnknownErrorCode, "/v2/project/stages/blobs/sha256:aaaa"), false),
	ginkgo.Entry("another tag is missing", registryErrorAt(transport.ManifestUnknownErrorCode, "/v2/project/stages/manifests/other-marker"), false),
	ginkgo.Entry("the tag request is denied", registryErrorAt(transport.DeniedErrorCode, "/v2/project/stages/manifests/marker"), false),
	ginkgo.Entry("the error carries no request", errors.New("MANIFEST_UNKNOWN: manifest unknown"), false),
	ginkgo.Entry("no error", nil, false),
)
