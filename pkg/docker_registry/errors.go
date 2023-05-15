package docker_registry

import (
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

func IsHarborNotFoundError(err error) bool {
	return (err != nil) && strings.Contains(err.Error(), "NOT_FOUND")
}
