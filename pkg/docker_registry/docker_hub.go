package docker_registry

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/flant/werf/pkg/image"
)

const DockerHubImplementationName = "dockerhub"

type DockerHubUnauthorizedError apiError

type DockerHubNotFoundError apiError

var dockerHubPatterns = []string{"^index\\.docker\\.io", "^registry\\.hub\\.docker\\.com"}

type dockerHub struct {
	*defaultImplementation
	dockerHubApi
	dockerHubCredentials

	token string
}

type dockerHubOptions struct {
	defaultImplementationOptions
	dockerHubCredentials
}

type dockerHubCredentials struct {
	username string
	password string
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

func (r *dockerHub) DeleteRepo(reference string) error {
	return r.deleteRepo(reference, true)
}

func (r *dockerHub) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := r.deleteRepoImage(repoImage, true); err != nil {
			return err
		}
	}

	return nil
}

func (r *dockerHub) deleteRepo(reference string, withRetry bool) error {
	token, err := r.getToken()
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	resp, err := r.dockerHubApi.deleteRepository(account, project, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized && withRetry {
			r.resetToken()
			return r.deleteRepo(reference, false)
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError{error: err}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *dockerHub) deleteRepoImage(repoImage *image.Info, withRetry bool) error {
	token, err := r.getToken()
	if err != nil {
		return err
	}

	account, project, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	resp, err := r.dockerHubApi.deleteTag(account, project, repoImage.Tag, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized && withRetry {
			r.resetToken()
			return r.deleteRepoImage(repoImage, false)
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError{error: err}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *dockerHub) getToken() (string, error) {
	if r.token == "" {
		token, resp, err := r.dockerHubApi.getToken(r.dockerHubCredentials.username, r.dockerHubCredentials.password)
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

		r.token = token
	}

	return r.token, nil
}

func (r *dockerHub) resetToken() {
	r.token = ""
}

func (r *dockerHub) Validate() error {
	if _, err := r.getToken(); err != nil {
		return err
	}

	return nil
}

func (r *dockerHub) String() string {
	return DockerHubImplementationName
}

func (r *dockerHub) parseReference(reference string) (string, string, error) {
	parsedReference, err := name.NewTag(reference)
	if err != nil {
		return "", "", err
	}

	repository := parsedReference.RepositoryStr()

	var account, project string
	switch len(strings.Split(repository, "/")) {
	case 1:
		account = repository
	case 2:
		project = path.Base(repository)
		account = path.Base(strings.TrimSuffix(repository, project))
	default:
		return "", "", fmt.Errorf("unexpected reference %s", reference)
	}

	return account, project, nil
}
