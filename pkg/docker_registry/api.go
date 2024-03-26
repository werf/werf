package docker_registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
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

func (api *api) tryGetRepoImage(ctx context.Context, reference, implementation string) (*image.Info, error) {
	if imgInfo, err := api.GetRepoImage(ctx, reference); err != nil {
		if implementation != "" {
			if IsQuayTagExpiredErr(err) && implementation != QuayImplementationName {
				logboek.Context(ctx).Error().LogF("WARNING: Detected error specific for quay container registry implementation!\n")
				logboek.Context(ctx).Error().LogF("WARNING: Use --repo-container-registry=quay option (or WERF_CONTAINER_REGISTRY env var)\n")
				logboek.Context(ctx).Error().LogF("WARNING:  to instruct werf to use quay driver.\n")
			}
		}
		if IsImageNotFoundError(err) || IsBrokenImageError(err) {
			// TODO: 1. make sure werf never ever creates rejected image records for name-unknown errors.
			// TODO: 2. werf-cleanup should remove broken images
			if debugDockerRegistry() {
				logboek.Context(ctx).Error().LogF("WARNING: Got an error when inspecting repo image %q: %s\n", reference, err)
			}
			return nil, nil
		}
		return imgInfo, err
	} else {
		return imgInfo, nil
	}
}

func (api *api) TryGetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	return api.tryGetRepoImage(ctx, reference, "")
}

func (api *api) GetRepoImage(ctx context.Context, reference string) (*image.Info, error) {
	desc, ref, err := api.getImageDesc(ctx, reference)
	if err != nil {
		return nil, err
	}

	referenceParts, err := api.parseReferenceParts(ref.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to parse reference %q: %w", ref.Name(), err)
	}

	return api.getRepoImageByDesc(ctx, referenceParts.tag, desc, ref)
}

func (api *api) getRepoImageByDesc(ctx context.Context, originalTag string, desc *remote.Descriptor, ref name.Reference) (*image.Info, error) {
	referenceParts, err := api.parseReferenceParts(ref.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to parse reference %q: %w", ref.Name(), err)
	}

	repoImage := &image.Info{
		Name:       ref.Name(),
		Repository: strings.Join([]string{referenceParts.registry, referenceParts.repository}, "/"),
		Tag:        originalTag,
	}

	if desc.MediaType.IsIndex() {
		repoImage.IsIndex = true

		ii, err := desc.ImageIndex()
		if err != nil {
			return nil, fmt.Errorf("error getting image index: %w", err)
		}

		digest, err := ii.Digest()
		if err != nil {
			return nil, fmt.Errorf("error getting image index digest: %w", err)
		}
		repoImage.RepoDigest = fmt.Sprintf("%s@%s", image.NormalizeRepository(repoImage.Repository), digest.String())

		im, err := ii.IndexManifest()
		if err != nil {
			return nil, fmt.Errorf("error getting image manifest: %w", err)
		}

		for _, desc := range im.Manifests {
			subref := fmt.Sprintf("%s@%s", repoImage.Repository, desc.Digest)
			subdesc, r, err := api.getImageDesc(ctx, subref)
			if err != nil {
				return nil, fmt.Errorf("error getting image %s manifest: %w", subref, err)
			}

			subInfo, err := api.getRepoImageByDesc(ctx, originalTag, subdesc, r)
			if err != nil {
				return nil, fmt.Errorf("error getting image %s descriptor: %w", subref, err)
			}
			repoImage.Index = append(repoImage.Index, subInfo)
		}
	} else {
		imageInfo, err := desc.Image()
		if err != nil {
			return nil, fmt.Errorf("error getting image manifest: %w", err)
		}

		digest, err := imageInfo.Digest()
		if err != nil {
			return nil, err
		}
		repoImage.RepoDigest = fmt.Sprintf("%s@%s", image.NormalizeRepository(repoImage.Repository), digest.String())

		manifest, err := imageInfo.Manifest()
		if err != nil {
			return nil, err
		}
		repoImage.ID = manifest.Config.Digest.String()

		configFile, err := imageInfo.ConfigFile()
		if err != nil {
			return nil, err
		}
		repoImage.Labels = configFile.Config.Labels
		repoImage.OnBuild = configFile.Config.OnBuild
		repoImage.Env = configFile.Config.Env
		repoImage.SetCreatedAtUnix(configFile.Created.Unix())

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
		repoImage.Size = totalSize

		parentID := configFile.Config.Image
		if parentID == "" {
			if id, ok := configFile.Config.Labels[image.WerfBaseImageIDLabel]; ok { // built with werf and buildah backend
				parentID = id
			}
		}
		repoImage.ParentID = parentID
	}

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

func (api *api) mutateImageOrIndex(ctx context.Context, imageOrIndex interface{}, mutateManifestConfigFunc func(cfg v1.Config) (v1.Config, error), ref name.Reference, isRefByDigest bool) (interface{}, error) {
	switch i := imageOrIndex.(type) {
	case v1.Image:
		cf, err := i.ConfigFile()
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		newconf, err := mutateManifestConfigFunc(cf.Config)
		if err != nil {
			return nil, err
		}
		newimg, err := mutate.Config(i, newconf)
		if err != nil {
			return nil, fmt.Errorf("unable to mutate image config: %w", err)
		}
		newdigest, err := newimg.Digest()
		if err != nil {
			return nil, fmt.Errorf("unable to get new image digest: %w", err)
		}
		var destref name.Reference
		if isRefByDigest {
			destref = ref.Context().Digest(newdigest.String())
		} else {
			destref = ref
		}
		if err := api.writeToRemote(ctx, destref, newimg); err != nil {
			return nil, fmt.Errorf("unable to write image: %w", err)
		}
		return newimg, nil
	case v1.ImageIndex:
		indexManifest, err := i.IndexManifest()
		if err != nil {
			return nil, fmt.Errorf("getting image index manifest: %w", err)
		}

		var newii v1.ImageIndex
		newii = empty.Index
		newii = mutate.IndexMediaType(newii, types.DockerManifestList)
		newiiadds := make([]mutate.IndexAddendum, 0, len(indexManifest.Manifests))

		for _, desc := range indexManifest.Manifests {
			subref := ref.Context().Digest(desc.Digest.String())
			switch {
			case desc.MediaType.IsIndex():
				subii, err := i.ImageIndex(desc.Digest)
				if err != nil {
					return nil, fmt.Errorf("getting manifest list index by digest %s: %w", desc.Digest, err)
				}
				var newsubii v1.ImageIndex
				if ret, err := api.mutateImageOrIndex(ctx, subii, mutateManifestConfigFunc, subref, true); err != nil {
					return nil, fmt.Errorf("unable to mutate index %s from manifest list: %w", subref, err)
				} else {
					newsubii = ret.(v1.ImageIndex)
				}
				newdesc, err := partial.Descriptor(newsubii)
				if err != nil {
					return nil, fmt.Errorf("unable to create image index descriptor of %s: %w", subref, err)
				}

				newiiadds = append(newiiadds, mutate.IndexAddendum{
					Add:        newsubii,
					Descriptor: *newdesc,
				})
			case desc.MediaType.IsImage():
				subimg, err := i.Image(desc.Digest)
				if err != nil {
					return nil, fmt.Errorf("getting image by digest %s: %w", desc.Digest, err)
				}
				var newsubimg v1.Image
				if ret, err := api.mutateImageOrIndex(ctx, subimg, mutateManifestConfigFunc, subref, true); err != nil {
					return nil, fmt.Errorf("unable to mutate image %s from manifest list: %w", subref, err)
				} else {
					newsubimg = ret.(v1.Image)
				}
				newsubcf, err := newsubimg.ConfigFile()
				if err != nil {
					return nil, fmt.Errorf("unable to get config file of %s: %w", subref, err)
				}
				newdesc, err := partial.Descriptor(newsubimg)
				if err != nil {
					return nil, fmt.Errorf("unable to create image descriptor of %q: %w", subref, err)
				}
				newdesc.Platform = newsubcf.Platform()

				newiiadds = append(newiiadds, mutate.IndexAddendum{
					Add:        newsubimg,
					Descriptor: *newdesc,
				})
			default:
				return nil, fmt.Errorf("unsupported media type %q: %w", desc.MediaType, err)
			}
		}

		newii = mutate.AppendManifests(newii, newiiadds...)
		newdigest, err := newii.Digest()
		if err != nil {
			return nil, fmt.Errorf("error getting new image index digest: %w", err)
		}
		var destref name.Reference
		if isRefByDigest {
			destref = ref.Context().Digest(newdigest.String())
		} else {
			destref = ref
		}
		if err := api.writeToRemote(ctx, destref, newii); err != nil {
			return nil, fmt.Errorf("unable to write index: %w", err)
		}
		return newii, nil
	default:
		panic("unexpected condition")
	}
}

func (api *api) MutateAndPushImage(ctx context.Context, sourceReference, destinationReference string, mutateManifestConfigFunc func(cfg v1.Config) (v1.Config, error)) error {
	dstRef, err := name.ParseReference(destinationReference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", destinationReference, err)
	}
	_, isDstDigest := dstRef.(name.Digest)

	desc, _, err := api.getImageDesc(ctx, sourceReference)
	if err != nil {
		return fmt.Errorf("error reading image %q: %w", sourceReference, err)
	}
	switch {
	case desc.MediaType.IsIndex():
		ii, err := desc.ImageIndex()
		if err != nil {
			return fmt.Errorf("getting image index: %w", err)
		}
		_, err = api.mutateImageOrIndex(ctx, ii, mutateManifestConfigFunc, dstRef, isDstDigest)
		return err
	case desc.MediaType.IsImage():
		img, err := desc.Image()
		if err != nil {
			return fmt.Errorf("error getting image manifest: %w", err)
		}
		_, err = api.mutateImageOrIndex(ctx, img, mutateManifestConfigFunc, dstRef, isDstDigest)
		return err
	default:
		return fmt.Errorf("unsupported media type %q: %w", desc.MediaType, err)
	}
}

func (api *api) CopyImage(ctx context.Context, sourceReference, destinationReference string, opts CopyImageOptions) error {
	dstRef, err := name.ParseReference(destinationReference, api.parseReferenceOptions()...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", destinationReference, err)
	}
	desc, _, err := api.getImageDesc(ctx, sourceReference)
	if err != nil {
		return fmt.Errorf("unable to get image %s: %w", sourceReference, err)
	}
	switch {
	case desc.MediaType.IsIndex():
		ii, err := desc.ImageIndex()
		if err != nil {
			return fmt.Errorf("getting image index %s: %w", sourceReference, err)
		}
		digest, err := ii.Digest()
		if err != nil {
			return fmt.Errorf("getting image index %s digest: %w", sourceReference, err)
		}
		logboek.Context(ctx).Debug().LogF("-- CopyImage writing index ref:%q digest:%q\n", dstRef, digest)
		if err := api.writeToRemote(ctx, dstRef, ii); err != nil {
			return fmt.Errorf("unable to write %s: %w", dstRef, err)
		}
	case desc.MediaType.IsImage():
		img, err := desc.Image()
		if err != nil {
			return fmt.Errorf("getting image manifest %s: %w", sourceReference, err)
		}
		digest, err := img.Digest()
		if err != nil {
			return fmt.Errorf("getting image %s digest: %w", sourceReference, err)
		}
		logboek.Context(ctx).Debug().LogF("-- CopyImage writing image ref:%q digest:%q\n", dstRef, digest)
		if err := api.writeToRemote(ctx, dstRef, img); err != nil {
			return fmt.Errorf("unable to write %s: %w", dstRef, err)
		}
	default:
		return fmt.Errorf("unsupported media type %q: %w", desc.MediaType, err)
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

func (api *api) getImageDesc(ctx context.Context, reference string) (*remote.Descriptor, name.Reference, error) {
	ref, err := name.ParseReference(reference, api.parseReferenceOptions()...)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reference %q: %w", reference, err)
	}

	desc, err := remote.Get(ref, api.defaultRemoteOptions(ctx)...)
	if err != nil {
		return nil, nil, fmt.Errorf("getting %s: %w", ref, err)
	}
	return desc, ref, nil
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
		remote.WithTransport(getHttpTransport(api.SkipTlsVerifyRegistry)),
	}
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

func (api *api) writeToRemote(ctx context.Context, ref name.Reference, imageOrIndex interface{}) error {
	c := make(chan v1.Update, 200)

	switch i := imageOrIndex.(type) {
	case v1.Image:
		go remote.Write(
			ref, i,
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithProgress(c),
			remote.WithTransport(getHttpTransport(api.SkipTlsVerifyRegistry)),
			remote.WithContext(ctx),
		)
	case v1.ImageIndex:
		go remote.WriteIndex(
			ref, i,
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithProgress(c),
			remote.WithTransport(getHttpTransport(api.SkipTlsVerifyRegistry)),
			remote.WithContext(ctx),
		)
	default:
		panic(fmt.Sprintf("unexpected object type %#v", i))
	}

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
		newdesc, err := partial.Descriptor(img)
		if err != nil {
			return fmt.Errorf("unable to create image descriptor of %q: %w", info.Name, err)
		}
		newdesc.Platform = cf.Platform()

		adds = append(adds, mutate.IndexAddendum{
			Add:        img,
			Descriptor: *newdesc,
		})
	}

	// TODO(multiarch): research whether we could use digest-only manifest writing mode
	// digest, err := ii.Digest()
	// if err != nil {
	// 	return fmt.Errorf("unable to calculate manifest list digest: %w", err)
	// }
	//  if _, ok := ref.(name.Digest); ok {
	// 		ref = ref.Context().Digest(digest.String())
	// }

	ii = mutate.AppendManifests(ii, adds...)

	digest, err := ii.Digest()
	if err != nil {
		return fmt.Errorf("error getting image index digest: %w", err)
	}
	logboek.Context(ctx).Debug().LogF("-- PushManifestList ref:%q digest:%q\n", manifestListRef, digest)

	if err := api.writeToRemote(ctx, manifestListRef, ii); err != nil {
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
