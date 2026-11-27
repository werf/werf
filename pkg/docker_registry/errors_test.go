package docker_registry

import (
	"errors"
	"net/http"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("IsTagNotFoundError", func(err error, expected bool) {
	gomega.Expect(IsTagNotFoundError(err, "marker")).To(gomega.Equal(expected))
},
	ginkgo.Entry("the tag is missing", registryErrorAt(http.StatusNotFound, transport.ManifestUnknownErrorCode, "manifest unknown", "/v2/project/stages/manifests/marker"), true),
	ginkgo.Entry("the repository is missing", registryErrorAt(http.StatusNotFound, transport.NameUnknownErrorCode, "repository name not known", "/v2/project/stages/manifests/marker"), true),
	ginkgo.Entry("the tag is missing in Harbor", registryErrorAt(http.StatusNotFound, "NOT_FOUND", "artifact not found", "/v2/project/stages/manifests/marker"), true),
	ginkgo.Entry("a child manifest of the found tag is missing", registryErrorAt(http.StatusNotFound, transport.ManifestUnknownErrorCode, "manifest unknown", "/v2/project/stages/manifests/sha256:aaaa"), false),
	ginkgo.Entry("a blob of the found tag is missing", registryErrorAt(http.StatusNotFound, transport.BlobUnknownErrorCode, "blob unknown", "/v2/project/stages/blobs/sha256:aaaa"), false),
	ginkgo.Entry("another tag is missing", registryErrorAt(http.StatusNotFound, transport.ManifestUnknownErrorCode, "manifest unknown", "/v2/project/stages/manifests/other-marker"), false),
	ginkgo.Entry("the tag request is denied", registryErrorAt(http.StatusForbidden, transport.DeniedErrorCode, "read access denied", "/v2/project/stages/manifests/marker"), false),
	ginkgo.Entry("the denial message mentions a not-found code", registryErrorAt(http.StatusForbidden, transport.DeniedErrorCode, "permission lookup failed: NOT_FOUND", "/v2/project/stages/manifests/marker"), false),
	ginkgo.Entry("a not-found status carries another code", registryErrorAt(http.StatusNotFound, transport.DeniedErrorCode, "upstream returned MANIFEST_UNKNOWN", "/v2/project/stages/manifests/marker"), false),
	ginkgo.Entry("the registry fails with a not-found code", registryErrorAt(http.StatusInternalServerError, transport.ManifestUnknownErrorCode, "upstream registry failure", "/v2/project/stages/manifests/marker"), false),
	ginkgo.Entry("the error carries no request", errors.New("MANIFEST_UNKNOWN: manifest unknown"), false),
	ginkgo.Entry("no error", nil, false),
)
