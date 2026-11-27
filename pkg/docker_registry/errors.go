package docker_registry

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

var (
	BrokenImageCodes = []transport.ErrorCode{
		transport.BlobUnknownErrorCode,
		transport.BlobUploadInvalidErrorCode,
		transport.BlobUploadUnknownErrorCode,
		transport.DigestInvalidErrorCode,
		transport.ManifestBlobUnknownErrorCode,
		transport.ManifestInvalidErrorCode,
		transport.ManifestUnverifiedErrorCode,
		transport.NameInvalidErrorCode,
	}

	NotFoundImageCodes = []transport.ErrorCode{
		transport.ManifestUnknownErrorCode,
		transport.NameUnknownErrorCode,
	}
)

func isErrorMatched(err error, codes []transport.ErrorCode) bool {
	if err != nil {
		for _, code := range codes {
			if strings.Contains(err.Error(), string(code)) {
				return true
			}
		}
	}
	return false
}

func IsBrokenImageError(err error) bool {
	return isErrorMatched(err, BrokenImageCodes) || IsQuayTagExpiredErr(err)
}

func IsImageNotFoundError(err error) bool {
	return isErrorMatched(err, NotFoundImageCodes) || IsHarborNotFoundError(err)
}

// Harbor answers NOT_FOUND where the distribution spec says MANIFEST_UNKNOWN.
var tagNotFoundCodes = []transport.ErrorCode{
	transport.ManifestUnknownErrorCode,
	transport.NameUnknownErrorCode,
	"NOT_FOUND",
}

// IsTagNotFoundError reports that the registry answered the request for the tag itself with 404
// and a not-found code. A found tag whose child manifest or blob is missing does not match.
func IsTagNotFoundError(err error, tag string) bool {
	var transportErr *transport.Error
	if !errors.As(err, &transportErr) || transportErr.StatusCode != http.StatusNotFound || transportErr.Request == nil {
		return false
	}
	if !strings.HasSuffix(transportErr.Request.URL.Path, "/manifests/"+tag) {
		return false
	}
	return slices.ContainsFunc(transportErr.Errors, func(diagnostic transport.Diagnostic) bool {
		return slices.Contains(tagNotFoundCodes, diagnostic.Code)
	})
}

func IsHarborNotFoundError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "NOT_FOUND")
}
