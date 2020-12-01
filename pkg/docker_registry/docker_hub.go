package docker_registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/werf/pkg/image"
)

const DockerHubImplementationName = "dockerhub"

type DockerHubUnauthorizedError apiError

type DockerHubNotFoundError apiError

var dockerHubPatterns = []string{"^index\\.docker\\.io", "^registry\\.hub\\.docker\\.com"}

type dockerHub struct {
	*defaultImplementation
	dockerHubApi
	dockerHubCredentials
}

type dockerHubOptions struct {
	defaultImplementationOptions
	dockerHubCredentials
}

type dockerHubCredentials struct {
	username string
	password string
	token    string
}

func newDockerHub(options dockerHubOptions) (*dockerHub, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	dockerHub := &dockerHub{
		defaultImplementation: d,
		dockerHubApi:          newDockerHubApi(),
		dockerHubCredentials:  options.dockerHubCredentials,
	}

	return dockerHub, nil
}

func (r *dockerHub) DeleteRepo(ctx context.Context, reference string) error {
	return r.deleteRepo(ctx, reference)
}

func (r *dockerHub) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	resp, err := r.dockerHubApi.deleteTag(ctx, account, project, repoImage.Tag, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return DockerHubUnauthorizedError{error: err}
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError{error: err}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *dockerHub) deleteRepo(ctx context.Context, reference string) error {
	token, err := r.getToken(ctx)
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.dockerHubApi.deleteRepository(ctx, account, project, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return DockerHubUnauthorizedError{error: err}
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError{error: err}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *dockerHub) getToken(ctx context.Context) (string, error) {
	if r.dockerHubCredentials.token == "" {
		token, resp, err := r.dockerHubApi.getToken(ctx, r.dockerHubCredentials.username, r.dockerHubCredentials.password)
		if resp != nil {
			if resp.StatusCode == http.StatusUnauthorized {
				return "", DockerHubUnauthorizedError{error: err}
			} else if resp.StatusCode == http.StatusNotFound {
				return "", DockerHubNotFoundError{error: err}
			}
		}

		if err != nil {
			return "", err
		}

		r.dockerHubCredentials.token = token
	}

	return r.dockerHubCredentials.token, nil
}

func (r *dockerHub) String() string {
	return DockerHubImplementationName
}

func (r *dockerHub) parseReference(reference string) (string, string, error) {
	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", err
	}

	repositoryParts := strings.Split(parsedReference.RepositoryStr(), "/")

	var account, project string
	if len(repositoryParts) == 2 {
		account = repositoryParts[0]
		project = repositoryParts[1]
	} else {
		return "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	return account, project, nil
}
