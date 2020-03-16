package docker_registry

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/flant/werf/pkg/image"
)

type GitLab struct {
	*Default
}

func (g *GitLab) DeleteRepoImage(repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if err := g.deleteRepoImage(repoImage); err != nil {
			return err
		}
	}

	return nil
}

func (g *GitLab) deleteRepoImage(repoImage *image.Info) error {
	if err := g.Default.DeleteRepoImage(repoImage); err != nil {
		if strings.Contains(err.Error(), "UNAUTHORIZED") {
			reference := strings.Join([]string{repoImage.Repository, repoImage.Digest}, "@")
			if secondDeleteErr := g.deleteRepoImageWithAllScopes(reference); secondDeleteErr != nil {
				if strings.Contains(secondDeleteErr.Error(), "UNAUTHORIZED") {
					return err
				}

				return secondDeleteErr
			}

			return nil
		}

		return err
	}

	return nil
}

// TODO https://gitlab.com/gitlab-org/gitlab-ce/issues/48968
func (g *GitLab) deleteRepoImageWithAllScopes(reference string) error {
	ref, err := name.ParseReference(reference, g.api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", reference, err)
	}

	auth, authErr := authn.DefaultKeychain.Resolve(ref.Context().Registry)
	if authErr != nil {
		return fmt.Errorf("getting creds for %q: %v", ref, authErr)
	}

	scopes := []string{ref.Scope("*")}
	tr, err := transport.New(ref.Context().Registry, auth, g.api.getHttpTransport(), scopes)
	if err != nil {
		return err
	}
	c := &http.Client{Transport: tr}

	u := url.URL{
		Scheme: ref.Context().Registry.Scheme(),
		Host:   ref.Context().RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s/manifests/%s", ref.Context().RepositoryStr(), ref.Identifier()),
	}

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		return nil
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unrecognized status code during DELETE: %v; %v", resp.Status, string(b))
	}
}
