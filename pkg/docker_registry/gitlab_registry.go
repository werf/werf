package docker_registry

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

const GitLabRegistryImplementationName = "gitlab"

var (
	gitlabPatterns = []string{`^gitlab\.com`}

	fullScopeFunc = func(ref name.Reference) []string {
		completeScopeFunc := []string{ref.Scope("push"), ref.Scope("pull"), ref.Scope("delete")}
		return completeScopeFunc
	}

	universalScopeFunc = func(ref name.Reference) []string {
		return []string{ref.Scope("*")}
	}
)

type gitLabRegistry struct {
	*defaultImplementation
	deleteRepoImageFunc func(ctx context.Context, repoImage *image.Info) error
}

type gitLabRegistryOptions struct {
	defaultImplementationOptions
}

func newGitLabRegistry(options gitLabRegistryOptions) (*gitLabRegistry, error) {
	d, err := newDefaultAPIForImplementation(GitLabRegistryImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	gitLab := &gitLabRegistry{defaultImplementation: d}

	return gitLab, nil
}

func (r *gitLabRegistry) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	if r.deleteRepoImageFunc != nil {
		return r.deleteRepoImageFunc(ctx, repoImage)
	}

	// DELETE /v2/<name>/tags/reference/<reference> method is available since the v2.8.0-gitlab
	var err error
	for _, deleteFunc := range []func(ctx context.Context, repoImage *image.Info) error{
		r.deleteRepoImageTagWithFullScope,
		r.deleteRepoImageTagWithUniversalScope,
	} {
		if err := deleteFunc(ctx, repoImage); err != nil {
			reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
			if strings.Contains(err.Error(), "404 Not Found; 404 page not found") {
				logboek.Context(ctx).Debug().LogF("DEBUG: %s: %s", reference, err)
				break
			} else if strings.Contains(err.Error(), "UNAUTHORIZED") {
				logboek.Context(ctx).Debug().LogF("DEBUG: %s: %s", reference, err)
				continue
			}

			return err
		}

		r.deleteRepoImageFunc = deleteFunc
		return nil
	}

	for _, deleteFunc := range []func(ctx context.Context, repoImage *image.Info) error{
		r.deleteRepoImageWithFullScope,
		r.deleteRepoImageWithUniversalScope,
	} {
		if err := deleteFunc(ctx, repoImage); err != nil {
			reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
			if strings.Contains(err.Error(), "UNAUTHORIZED") {
				logboek.Context(ctx).Debug().LogF("DEBUG: %s: %s", reference, err)
				continue
			}

			return err
		}

		r.deleteRepoImageFunc = deleteFunc
		return nil
	}

	err = r.defaultImplementation.DeleteRepoImage(ctx, repoImage)
	if err != nil {
		return err
	}

	r.deleteRepoImageFunc = r.defaultImplementation.DeleteRepoImage
	return nil
}

func (r *gitLabRegistry) deleteRepoImageTagWithUniversalScope(_ context.Context, repoImage *image.Info) error {
	return r.deleteRepoImageTagWithCustomScope(repoImage, universalScopeFunc)
}

func (r *gitLabRegistry) deleteRepoImageTagWithFullScope(_ context.Context, repoImage *image.Info) error {
	return r.deleteRepoImageTagWithCustomScope(repoImage, fullScopeFunc)
}

func (r *gitLabRegistry) deleteRepoImageWithUniversalScope(_ context.Context, repoImage *image.Info) error {
	return r.deleteRepoImageWithCustomScope(repoImage, universalScopeFunc)
}

func (r *gitLabRegistry) deleteRepoImageWithFullScope(_ context.Context, repoImage *image.Info) error {
	return r.deleteRepoImageWithCustomScope(repoImage, fullScopeFunc)
}

func (r *gitLabRegistry) deleteRepoImageTagWithCustomScope(repoImage *image.Info, scopeFunc func(ref name.Reference) []string) error {
	reference := strings.Join([]string{repoImage.Repository, repoImage.Tag}, ":")
	return r.customDeleteRepoImage("/v2/%s/tags/reference/%s", reference, scopeFunc)
}

func (r *gitLabRegistry) deleteRepoImageWithCustomScope(repoImage *image.Info, scopeFunc func(ref name.Reference) []string) error {
	return r.customDeleteRepoImage("/v2/%s/manifests/%s", repoImage.RepoDigest, scopeFunc)
}

func (r *gitLabRegistry) customDeleteRepoImage(endpointFormat, reference string, scopeFunc func(ref name.Reference) []string) error {
	ref, err := name.ParseReference(reference, r.api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	auth, authErr := authn.DefaultKeychain.Resolve(ref.Context().Registry)
	if authErr != nil {
		return fmt.Errorf("getting creds for %q: %w", ref, authErr)
	}

	scope := scopeFunc(ref)
	tr, err := transport.New(ref.Context().Registry, auth, getHttpTransport(false), scope)
	if err != nil {
		return err
	}
	c := &http.Client{Transport: tr}

	u := url.URL{
		Scheme: ref.Context().Registry.Scheme(),
		Host:   ref.Context().RegistryStr(),
		Path:   fmt.Sprintf(endpointFormat, ref.Context().RepositoryStr(), ref.Identifier()),
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

func (r *gitLabRegistry) String() string {
	return GitLabRegistryImplementationName
}
