package docker_registry

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	dockerReference "github.com/docker/distribution/reference"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker_registry/container_registry_extensions"
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

func (api *api) Tags(ctx context.Context, reference string, _ ...Option) ([]string, error) {
	return api.tags(ctx, reference)
}

func (api *api) tags(ctx context.Context, reference string, extraListOptions ...remote.Option) ([]string, error) {
	tags, err := api.list(ctx, reference, extraListOptions...)
	if err != nil {
		if IsStatusNotFoundErr(err) {
			return []string{}, nil
		}

		return nil, err
	}

	return tags, nil
}

func (api *api) tagImage(ctx context.Context, reference, tag string) error {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	desc, err := remote.Get(ref, api.defaultRemoteOptions(ctx)...)
	if err != nil {
		return fmt.Errorf("getting reference %q: %w", reference, err)
	}

	dst := ref.Context().Tag(tag)

	return remote.Tag(dst, desc, api.defaultRemoteOptions(ctx)...)
}

func (api *api) IsRepoImageExists(ctx context.Context, reference string) (bool, error) {
	if imgInfo, err := api.TryGetRepoImage(ctx, reference); err != nil {
		return false, err
	} else {
		return imgInfo != nil, nil
	}
}

func (api *api) TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	if imgInfo, err := api.GetRepoImage(ctx, reference); err != nil {
		if IsStatusNotFoundErr(err) || IsQuayTagExpiredErr(err) {
			// TODO: 1. auto reject images with manifest-unknown or blob-unknown errors
			// TODO: 2. why TryGetRepoImage for rejected image records gives manifest-unknown errors?
			// TODO: 3. make sure werf never ever creates rejected image records for name-unknown errors.
			// TODO: 4. werf-cleanup should remove broken images

			if os.Getenv("WERF_DOCKER_REGISTRY_DEBUG") == "1" {
				logboek.Context(ctx).Error().LogF("WARNING: Got an error when inspecting repo image %q: %s\n", reference, err)
			}

			return nil, nil
		}

		return imgInfo, err
	} else {
		return imgInfo, nil
	}
}

func (api *api) GetRepoImage(_ context.Context, reference string) (*image.Info, error) {
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

	var totalSize int64
	if layers, err := imageInfo.Layers(); err != nil {
		return nil, err
	} else {
		for _, l := range layers {
			if lSize, err := l.Size(); err != nil {
				return nil, err
			} else {
				totalSize += lSize
			}
		}
	}

	referenceParts, err := api.parseReferenceParts(reference)
	if err != nil {
		return nil, fmt.Errorf("unable to parse reference %q: %w", reference, err)
	}

	var parentID string
	if baseImageID, ok := configFile.Config.Labels["werf.io/base-image-id"]; ok {
		parentID = baseImageID
	} else {
		// TODO(1.3): Legacy compatibility mode
		parentID = configFile.Config.Image
	}

	repoImage := &image.Info{
		Name:       reference,
		Repository: strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/"),
		ID:         manifest.Config.Digest.String(),
		Tag:        referenceParts.tag,
		RepoDigest: digest.String(),
		ParentID:   parentID,
		Labels:     configFile.Config.Labels,
		OnBuild:    configFile.Config.OnBuild,
		Env:        configFile.Config.Env,
		Size:       totalSize,
	}

	repoImage.SetCreatedAtUnix(configFile.Created.Unix())

	return repoImage, nil
}

func (api *api) list(ctx context.Context, reference string, extraListOptions ...remote.Option) ([]string, error) {
	repo, err := name.NewRepository(reference, api.newRepositoryOptions()...)
	if err != nil {
		return nil, fmt.Errorf("parsing repo %q: %w", reference, err)
	}

	listOptions := append(
		api.defaultRemoteOptions(ctx),
		extraListOptions...,
	)
	tags, err := remote.List(repo, listOptions...)
	if err != nil {
		return nil, fmt.Errorf("reading tags for %q: %w", repo, err)
	}

	return tags, nil
}

func (api *api) deleteImageByReference(ctx context.Context, reference string) error {
	r, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	if err := remote.Delete(r, api.defaultRemoteOptions(ctx)...); err != nil {
		return fmt.Errorf("deleting image %q: %w", r, err)
	}

	return nil
}

func (api *api) MutateAndPushImage(_ context.Context, sourceReference, destinationReference string, mutateConfigFunc func(cfg v1.Config) (v1.Config, error)) error {
	img, _, err := api.image(sourceReference)
	if err != nil {
		return err
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return err
	}

	newConf, err := mutateConfigFunc(cfgFile.Config)
	if err != nil {
		return err
	}

	newImg, err := mutate.Config(img, newConf)
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(destinationReference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", destinationReference, err)
	}

	if err = remote.Write(ref, newImg, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return err
	}

	return nil
}

type PushImageOptions struct {
	Labels map[string]string
}

func (api *api) PushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	return api.pushWithRetry(ctx, func() error {
		return api.pushImage(ctx, reference, opts)
	})
}

func (api *api) pushImage(ctx context.Context, reference string, opts *PushImageOptions) error {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	labels := map[string]string{}
	if opts != nil {
		labels = opts.Labels
	}
	img := container_registry_extensions.NewManifestOnlyImage(labels)

	if err := api.writeToRemote(ctx, ref, img); err != nil {
		return fmt.Errorf("write to the remote %s have failed: %w", ref.String(), err)
	}

	return nil
}

func (api *api) image(reference string) (v1.Image, name.Reference, error) {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	img, err := remote.Image(
		ref,
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(api.getHttpTransport()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("reading image %q: %w", ref, err)
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

func (api *api) defaultRemoteOptions(ctx context.Context) []remote.Option {
	return []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(api.getHttpTransport()),
	}
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

type referenceParts struct {
	registry   string
	repository string
	tag        string
	digest     string
}

func (api *api) parseReferenceParts(reference string) (referenceParts, error) {
	parsedReference, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return referenceParts{}, err
	}

	// res[0] full match
	// res[1] repository
	// res[2] tag
	// res[3] digest
	res := dockerReference.ReferenceRegexp.FindStringSubmatch(reference)
	if len(res) != 4 {
		panic(fmt.Sprintf("unexpected regexp find submatch result %v for reference %q (%d)", res, reference, len(res)))
	}

	referenceParts := referenceParts{}
	referenceParts.registry = parsedReference.Context().RegistryStr()
	referenceParts.repository = parsedReference.Context().RepositoryStr()
	referenceParts.tag = res[2]
	if referenceParts.tag == "" {
		referenceParts.tag = name.DefaultTag
	}
	referenceParts.digest = res[3]

	return referenceParts, nil
}

func (api *api) PushImageArchive(ctx context.Context, archiveOpener ArchiveOpener, reference string) error {
	tag, err := name.NewTag(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("unable to parse reference %q: %w", reference, err)
	}

	img, err := tarball.Image(archiveOpener.Open, nil)
	if err != nil {
		return fmt.Errorf("unable to open tarball image: %w", err)
	}

	return api.pushWithRetry(ctx, func() error {
		if err := api.writeToRemote(ctx, tag, img); err != nil {
			return fmt.Errorf("write to the remote %s have failed: %w", tag.String(), err)
		}
		return nil
	})
}

func (api *api) pushWithRetry(ctx context.Context, pusher func() error) error {
	const retriesLimit = 5
	var err error

attemptLoop:
	for attempt := 1; attempt <= retriesLimit; attempt++ {
		err = pusher()

		if err != nil {
			for _, substr := range []string{
				"REDACTED: UNKNOWN",
				"http2: server sent GOAWAY and closed the connection",
				"http2: Transport received Server's graceful shutdown GOAWAY",
			} {
				if strings.Contains(err.Error(), substr) {
					seconds := rand.Intn(5) + 1

					msg := fmt.Sprintf("Retrying publishing in %d seconds (%d/%d) ...\n", seconds, attempt, retriesLimit)
					logboek.Context(ctx).Warn().LogLn(msg)

					time.Sleep(time.Duration(seconds) * time.Second)
					continue attemptLoop
				}
			}

			return err
		}

		return nil
	}

	return fmt.Errorf("retries limit reached: %w", err)
}

func (api *api) writeToRemote(ctx context.Context, ref name.Reference, img v1.Image) error {
	c := make(chan v1.Update, 200)

	go remote.Write(
		ref, img,
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithProgress(c),
		remote.WithTransport(api.getHttpTransport()),
		remote.WithContext(ctx),
	)

	for upd := range c {
		switch {
		case upd.Error != nil && errors.Is(upd.Error, io.EOF):
			logboek.Context(ctx).Debug().LogF("(%d/%d) done pushing image %q\n", upd.Complete, upd.Total, ref.String())
			return nil
		case upd.Error != nil:
			return fmt.Errorf("error pushing image: %w", upd.Error)
		default:
			logboek.Context(ctx).Debug().LogF("(%d/%d) pushing image %s is in progress\n", upd.Complete, upd.Total, ref.String())
		}
	}

	return nil
}

func (api *api) PullImageArchive(ctx context.Context, archiveWriter io.Writer, reference string) error {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("unable to parse reference %q: %w", reference, err)
	}

	desc, err := remote.Get(ref, api.defaultRemoteOptions(ctx)...)
	if err != nil {
		return fmt.Errorf("getting reference %q: %w", reference, err)
	}

	img, err := desc.Image()
	if err != nil {
		return fmt.Errorf("unable to resolve image manifest for reference %q: %w", reference, err)
	}

	c := make(chan v1.Update, 200)

	go tarball.Write(ref, img, archiveWriter, tarball.WithProgress(c))

	for upd := range c {
		switch {
		case upd.Error != nil && errors.Is(upd.Error, io.EOF):
			logboek.Context(ctx).Debug().LogF("(%d/%d) done pulling image %s to archive\n", upd.Complete, upd.Total, reference)
			return nil
		case upd.Error != nil:
			return fmt.Errorf("error receiving image data: %w", upd.Error)
		default:
			logboek.Context(ctx).Debug().LogF("(%d/%d) pulling image %s is in progress\n", upd.Complete, upd.Total, reference)
		}
	}

	return nil
}

func (api *api) PushManifestList(ctx context.Context, reference string, opts ManifestListOptions) error {
	if len(opts.Manifests) == 0 {
		panic("unexpected empty manifests list")
	}

	manifestListRef, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("unable to parse reference %q: %w", reference, err)
	}

	// FIXME(multiarch): Check manifest list already exists and do not republish in this case.
	// FIXME(multiarch): If custom tag, then we should check by digest => we need to get each manifest from registry.
	// FIXME(multiarch): If checksum tag, then we could simply check existing.
	// FIXME(multiarch): ^^ This behaviour should be controlled by an option.

	var ii v1.ImageIndex
	ii = empty.Index
	ii = mutate.IndexMediaType(ii, types.DockerManifestList)

	adds := make([]mutate.IndexAddendum, 0, len(opts.Manifests))

	for _, info := range opts.Manifests {
		ref, err := name.ParseReference(info.Name, api.parseReferenceOptions()...)
		if err != nil {
			return fmt.Errorf("unable to parse reference %q: %w", info.Name, err)
		}
		// FIXME(multiarch): Optimize: do not get manifest, save v1.Image into the image.Info (optional)
		desc, err := remote.Get(ref, api.defaultRemoteOptions(ctx)...)
		if err != nil {
			return fmt.Errorf("unable to get manifest of %q: %w", info.Name, err)
		}
		img, err := desc.Image()
		if err != nil {
			return fmt.Errorf("unable to get image descriptor of %q: %w", info.Name, err)
		}
		cf, err := img.ConfigFile()
		if err != nil {
			return fmt.Errorf("unable to get config file of %q: %w", info.Name, err)
		}
		newDesc, err := partial.Descriptor(img)
		if err != nil {
			return fmt.Errorf("unable to create image descriptor of %q: %w", info.Name, err)
		}
		newDesc.Platform = cf.Platform()

		adds = append(adds, mutate.IndexAddendum{
			Add:        img,
			Descriptor: *newDesc,
		})
	}

	ii = mutate.AppendManifests(ii, adds...)

	// TODO(multiarch): research whether we could use digest-only manifest writing mode
	// digest, err := ii.Digest()
	// if err != nil {
	// 	return fmt.Errorf("unable to calculate manifest list digest: %w", err)
	// }
	//  if _, ok := ref.(name.Digest); ok {
	// 		ref = ref.Context().Digest(digest.String())
	// }

	if err := remote.WriteIndex(manifestListRef, ii, api.defaultRemoteOptions(ctx)...); err != nil {
		return fmt.Errorf("unable to write manifest %q: %w", manifestListRef, err)
	}

	return nil
}

func ValidateRepositoryReference(reference string) error {
	reg := regexp.MustCompile(`^` + dockerReference.NameRegexp.String() + `$`)
	if !reg.MatchString(reference) {
		return fmt.Errorf("invalid repository address %q", reference)
	}

	return nil
}

func IsStatusNotFoundErr(err error) bool {
	var transportError *transport.Error
	if errors.As(err, &transportError) && transportError.StatusCode == 404 {
		return true
	}

	return false
}
