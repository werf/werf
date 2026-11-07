package bundles

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"

	"github.com/werf/werf/v3/pkg/ref"
)

// normalizeRepo expands a Docker Hub short repository name, e.g. "user/app" or "app", into the full
// name, e.g. "docker.io/user/app" or "docker.io/library/app", as Docker does.
func normalizeRepo(repo string) (string, error) {
	named, err := reference.ParseNormalizedNamed(repo)
	if err != nil {
		return "", fmt.Errorf("parse repository %q: %w", repo, err)
	}

	return named.Name(), nil
}

// normalizeAddr expands a Docker Hub short repository name in a bundle address
// "[docker://]REPO[:TAG|@DIGEST]", keeping the rest of the address. An archive address is kept as is.
func normalizeAddr(addr string) (string, error) {
	if strings.HasPrefix(addr, ref.ArchiveSchema) {
		return addr, nil
	}

	schema := ""
	if strings.HasPrefix(addr, ref.RegistrySchema) {
		schema = ref.RegistrySchema
	}

	repo := strings.TrimPrefix(addr, schema)
	suffix := ""
	if i := strings.Index(repo, "@"); i >= 0 {
		repo, suffix = repo[:i], repo[i:]
	} else if i := strings.LastIndex(repo, ":"); i >= 0 && i > strings.LastIndex(repo, "/") {
		repo, suffix = repo[:i], repo[i:]
	}

	repo, err := normalizeRepo(repo)
	if err != nil {
		return "", err
	}

	return schema + repo + suffix, nil
}
