package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

const QuayImplementationName = "quay"

type QuayNotFoundError apiError

var quayPatterns = []string{"^quay\\.io", "^quay\\..*\\.com"}

type quay struct {
	*defaultImplementation
	quayApi
	quayCredentials
}

type quayOptions struct {
	defaultImplementationOptions
	quayCredentials
}

type quayCredentials struct {
	token string
}

func newQuay(options quayOptions) (*quay, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	quay := &quay{
		defaultImplementation: d,
		quayApi:               newQuayApi(),
		quayCredentials:       options.quayCredentials,
	}

	return quay, nil
}

func (r *quay) ResolveRepoMode(_ context.Context, registryOrRepositoryAddress, repoMode string) (string, error) {
	_, _, repository, err := r.parseReference(registryOrRepositoryAddress)
	if err != nil {
		return "", err
	}

	switch repoMode {
	case MonorepoRepoMode:
		if repository != "" {
			return MonorepoRepoMode, nil
		}

		return "", fmt.Errorf("docker registry implementation %[1]s and repo mode %[2]s cannot be used with %[4]s (add repository to address or use %[3]s repo mode)", r.String(), MonorepoRepoMode, MultirepoRepoMode, registryOrRepositoryAddress)
	case MultirepoRepoMode:
		if repository == "" {
			return MultirepoRepoMode, nil
		}

		return "", fmt.Errorf("docker registry implementation %[1]s and repo mode %[3]s cannot be used with %[4]s (exclude repository from address or use %[2]s repo mode)", r.String(), MonorepoRepoMode, MultirepoRepoMode, registryOrRepositoryAddress)
	case "auto", "":
		if repository == "" {
			return MultirepoRepoMode, nil
		} else {
			return MonorepoRepoMode, nil
		}
	default:
		return "", fmt.Errorf("docker registry implementation %s does not support repo mode %s", r.String(), repoMode)
	}
}

func (r *quay) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *quay) deleteRepo(ctx context.Context, reference string) error {
	hostname, namespace, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.quayApi.DeleteRepository(ctx, hostname, namespace, repository, r.quayCredentials.token)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return QuayNotFoundError{error: err}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *quay) String() string {
	return QuayImplementationName
}

func (r *quay) parseReference(reference string) (string, string, string, error) {
	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", "", err
	}

	hostname := parsedReference.RegistryStr()
	repositoryStr := parsedReference.RepositoryStr()

	var namespace, repository string
	switch len(strings.Split(repositoryStr, "/")) {
	case 1:
		namespace = repositoryStr
	case 2:
		repository = path.Base(repositoryStr)
		namespace = path.Base(strings.TrimSuffix(repositoryStr, repository))
	default:
		return "", "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	return hostname, namespace, repository, nil
}
