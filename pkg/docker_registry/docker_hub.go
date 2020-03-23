package docker_registry

import (
	"net/http"

	"github.com/flant/werf/pkg/image"
)

const DockerHubImplementationName = "dockerhub"

type DockerHubUnauthorizedError error
type DockerHubNotFoundError error

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

	resp, err := r.dockerHubApi.deleteRepository(reference, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized && withRetry {
			r.resetToken()
			return r.deleteRepo(reference, false)
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError(err)
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

	resp, err := r.dockerHubApi.deleteTag(repoImage.Repository, repoImage.Tag, token)
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized && withRetry {
			r.resetToken()
			return r.deleteRepoImage(repoImage, false)
		} else if resp.StatusCode == http.StatusNotFound {
			return DockerHubNotFoundError(err)
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
				return "", DockerHubUnauthorizedError(err)
			} else if resp.StatusCode == http.StatusNotFound {
				return "", DockerHubNotFoundError(err)
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
