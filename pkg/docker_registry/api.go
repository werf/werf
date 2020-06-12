package docker_registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/werf/werf/pkg/image"
)

type api struct {
	InsecureRegistry      bool
	SkipTlsVerifyRegistry bool
}

type apiOptions struct {
	InsecureRegistry      bool
	SkipTlsVerifyRegistry bool
}

func newAPI(options apiOptions) *api {
	return &api{
		InsecureRegistry:      options.InsecureRegistry,
		SkipTlsVerifyRegistry: options.SkipTlsVerifyRegistry,
	}
}

func (api *api) Tags(reference string) ([]string, error) {
	tags, err := api.list(reference)
	if err != nil {
		if IsNameUnknownError(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return tags, nil
}

func (api *api) IsRepoImageExists(reference string) (bool, error) {
	if imgInfo, err := api.TryGetRepoImage(reference); err != nil {
		return false, err
	} else {
		return imgInfo != nil, nil
	}
}

func (api *api) TryGetRepoImage(reference string) (*image.Info, error) {
	if imgInfo, err := api.GetRepoImage(reference); err != nil {
		if IsManifestUnknownError(err) || IsNameUnknownError(err) {
			return imgInfo, nil
		}
		return imgInfo, err
	} else {
		return imgInfo, nil
	}
}

func (api *api) GetRepoImage(reference string) (*image.Info, error) {
	imageInfo, _, err := api.image(reference)
	if err != nil {
		return nil, err
	}

	digest, err := imageInfo.Digest()
	if err != nil {
		return nil, err
	}

	manifest, err := imageInfo.Manifest()
	if err != nil {
		return nil, err
	}

	configFile, err := imageInfo.ConfigFile()
	if err != nil {
		return nil, err
	}

	parsedReference, err := name.NewTag(reference, api.parseReferenceOptions()...)
	if err != nil {
		return nil, err
	}

	repoImage := &image.Info{
		Name:       reference,
		Repository: strings.Join([]string{parsedReference.RegistryStr(), parsedReference.RepositoryStr()}, "/"),
		ID:         manifest.Config.Digest.String(),
		Tag:        parsedReference.TagStr(),
		RepoDigest: digest.String(),
		ParentID:   configFile.Config.Image,
		Labels:     configFile.Config.Labels,
	}

	repoImage.SetCreatedAtUnix(configFile.Created.Unix())

	return repoImage, nil
}

func (api *api) list(reference string) ([]string, error) {
	repo, err := name.NewRepository(reference, api.newRepositoryOptions()...)
	if err != nil {
		return nil, fmt.Errorf("parsing repo %q: %v", reference, err)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(authn.DefaultKeychain), remote.WithTransport(api.getHttpTransport()))
	if err != nil {
		return nil, fmt.Errorf("reading tags for %q: %v", repo, err)
	}

	return tags, nil
}

func (api *api) deleteImageByReference(reference string) error {
	r, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", reference, err)
	}

	if err := remote.Delete(r, remote.WithAuthFromKeychain(authn.DefaultKeychain), remote.WithTransport(api.getHttpTransport())); err != nil {
		return fmt.Errorf("deleting image %q: %v", r, err)
	}

	return nil
}

func (api *api) image(reference string) (v1.Image, name.Reference, error) {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reference %q: %v", reference, err)
	}

	// FIXME: Hack for the go-containerregistry library,
	// FIXME: that uses default transport without options to change transport to custom.
	// FIXME: Needed for the insecure https registry to work.
	oldDefaultTransport := http.DefaultTransport
	http.DefaultTransport = api.getHttpTransport()
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	http.DefaultTransport = oldDefaultTransport

	if err != nil {
		return nil, nil, fmt.Errorf("reading image %q: %v", ref, err)
	}

	return img, ref, nil
}

func (api *api) newRepositoryOptions() []name.Option {
	return api.parseReferenceOptions()
}

func (api *api) parseReferenceOptions() []name.Option {
	var options []name.Option
	options = append(options, name.WeakValidation)

	if api.InsecureRegistry {
		options = append(options, name.Insecure)
	}

	return options
}

func (api *api) getHttpTransport() (transport http.RoundTripper) {
	transport = http.DefaultTransport

	if api.SkipTlsVerifyRegistry {
		defaultTransport := http.DefaultTransport.(*http.Transport)

		newTransport := &http.Transport{
			Proxy:                 defaultTransport.Proxy,
			DialContext:           defaultTransport.DialContext,
			MaxIdleConns:          defaultTransport.MaxIdleConns,
			IdleConnTimeout:       defaultTransport.IdleConnTimeout,
			TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
			ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		}

		transport = newTransport
	}

	return
}
