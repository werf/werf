package docker_registry

import (
	"errors"
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

// IsTagNotFoundError tells a missing tag from a found tag whose child manifest or blob is missing.
func IsTagNotFoundError(err error, tag string) bool {
	var transportErr *transport.Error
	if !IsImageNotFoundError(err) || !errors.As(err, &transportErr) || transportErr.Request == nil {
		return false
	}
	return strings.HasSuffix(transportErr.Request.URL.Path, "/manifests/"+tag)
}

func IsHarborNotFoundError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "NOT_FOUND")
}
