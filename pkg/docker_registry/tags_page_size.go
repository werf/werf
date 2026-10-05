package docker_registry

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

const (
	tagsPageSizeEnv = "WERF_DOCKER_REGISTRY_TAGS_PAGE_SIZE"

	defaultTagsPageSize = 1_000_000
	libraryTagsPageSize = 0

	publicAwsEcrHost = "public.ecr.aws"

	paginationNumberInvalidErrorCode transport.ErrorCode = "PAGINATION_NUMBER_INVALID"
)

// ECR rejects pages larger than its own limit with an UNSUPPORTED diagnostic instead of the
// pagination error code: https://github.com/google/go-containerregistry/issues/681
var awsEcrMaxResultsRejectionRegexp = regexp.MustCompile(`(?i)parameter at 'maxresults'.*less than or equal to \d+`)

var libraryTagsPageSizeHosts sync.Map

func tagsPageSizeForRegistryHost(registryHost string, useLibraryTagsPageSize bool) (int, error) {
	pageSize, err := tagsPageSizeFromEnv()
	if err != nil {
		return 0, err
	}

	if useLibraryTagsPageSize || isPublicAwsEcrHost(registryHost) {
		return libraryTagsPageSize, nil
	}

	if _, capped := libraryTagsPageSizeHosts.Load(registryHost); capped {
		return libraryTagsPageSize, nil
	}

	return pageSize, nil
}

func tagsPageSizeFromEnv() (int, error) {
	value := strings.TrimSpace(os.Getenv(tagsPageSizeEnv))
	if value == "" {
		return defaultTagsPageSize, nil
	}

	pageSize, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("read %s: expected a non-negative integer number of tags per request", tagsPageSizeEnv)
	}
	if pageSize < 0 {
		return 0, fmt.Errorf("read %s: expected a non-negative integer number of tags per request, got %d", tagsPageSizeEnv, pageSize)
	}

	return pageSize, nil
}

func isPublicAwsEcrHost(registryHost string) bool {
	if hostname, _, err := net.SplitHostPort(registryHost); err == nil {
		return hostname == publicAwsEcrHost
	}

	return registryHost == publicAwsEcrHost
}

func isTagsPageSizeRejectedErr(err error) bool {
	var registryErr *transport.Error
	if !errors.As(err, &registryErr) {
		return false
	}

	switch registryErr.StatusCode {
	case http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnprocessableEntity:
	default:
		return false
	}

	if req := registryErr.Request; req != nil {
		if req.Method != http.MethodGet || !strings.HasSuffix(req.URL.Path, "/tags/list") {
			return false
		}
	}

	for _, diagnostic := range registryErr.Errors {
		if diagnostic.Code == paginationNumberInvalidErrorCode {
			return true
		}
		if awsEcrMaxResultsRejectionRegexp.MatchString(diagnostic.Message) {
			return true
		}
	}

	return false
}
