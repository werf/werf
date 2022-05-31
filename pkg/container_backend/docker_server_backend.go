package container_backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/image"
)

type DockerServerBackend struct{}

func NewDockerServerBackend() *DockerServerBackend {
	return &DockerServerBackend{}
}

func (runtime *DockerServerBackend) HasStapelBuildSupport() bool {
	return false
}

func (runtime *DockerServerBackend) BuildStapelStage(ctx context.Context, opts BuildStapelStageOptions) (string, error) {
	panic("BuildStapelStage does not implemented for DockerServerBackend. Please report the bug if you've received this message.")
}

func (runtime *DockerServerBackend) CalculateDependencyImportChecksum(ctx context.Context, dependencyImport DependencyImportSpec) (string, error) {
	panic("CalculateDependencyImportChecksum does not implemented for DockerServerBackend. Please report the bug if you've received this message.")
}

func (runtime *DockerServerBackend) BuildDockerfile(ctx context.Context, _ []byte, opts BuildDockerfileOpts) (string, error) {
	switch {
	case opts.ContextTar == nil:
		panic(fmt.Sprintf("ContextTar can't be nil: %+v", opts))
	case opts.DockerfileCtxRelPath == "":
		panic(fmt.Sprintf("DockerfileCtxRelPath can't be empty: %+v", opts))
	}

	var cliArgs []string

	cliArgs = append(cliArgs, "--file", opts.DockerfileCtxRelPath)
	if opts.Target != "" {
		cliArgs = append(cliArgs, "--target", opts.Target)
	}
	if opts.Network != "" {
		cliArgs = append(cliArgs, "--network", opts.Network)
	}
	if opts.SSH != "" {
		cliArgs = append(cliArgs, "--ssh", opts.SSH)
	}

	for _, addHost := range opts.AddHost {
		cliArgs = append(cliArgs, "--add-host", addHost)
	}
	for _, buildArg := range opts.BuildArgs {
		cliArgs = append(cliArgs, "--build-arg", buildArg)
	}
	for _, label := range opts.Labels {
		cliArgs = append(cliArgs, "--label", label)
	}

	tempID := uuid.New().String()
	opts.Tags = append(opts.Tags, tempID)
	for _, tag := range opts.Tags {
		cliArgs = append(cliArgs, "--tag", tag)
	}

	cliArgs = append(cliArgs, "-")

	if Debug() {
		fmt.Printf("[DOCKER BUILD] docker build %s\n", strings.Join(cliArgs, " "))
	}

	return tempID, docker.CliBuild_LiveOutputWithCustomIn(ctx, opts.ContextTar, cliArgs...)
}

// ShouldCleanupDockerfileImage for docker-server backend we should cleanup image built from dockerfrom tagged with tempID
// which is implementation detail of the BuildDockerfile.
func (runtime *DockerServerBackend) ShouldCleanupDockerfileImage() bool {
	return true
}

func (runtime *DockerServerBackend) GetImageInfo(ctx context.Context, ref string, opts GetImageInfoOpts) (*image.Info, error) {
	inspect, err := docker.ImageInspect(ctx, ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to inspect docker image: %w", err)
	}

	return image.NewInfoFromInspect(ref, inspect), nil
}

// GetImageInspect only available for DockerServerBackend
func (runtime *DockerServerBackend) GetImageInspect(ctx context.Context, ref string) (*types.ImageInspect, error) {
	inspect, err := docker.ImageInspect(ctx, ref)
	if client.IsErrNotFound(err) {
		return nil, nil
	}
	return inspect, err
}

func (runtime *DockerServerBackend) RefreshImageObject(ctx context.Context, img LegacyImageInterface) error {
	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}
	return nil
}

func (runtime *DockerServerBackend) RenameImage(ctx context.Context, img LegacyImageInterface, newImageName string, removeOldName bool) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging image %s by name %s", img.Name(), newImageName)).DoError(func() error {
		if err := docker.CliTag(ctx, img.Name(), newImageName); err != nil {
			return fmt.Errorf("unable to tag image %s by name %s: %w", img.Name(), newImageName, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if removeOldName {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing old image tag %s", img.Name())).DoError(func() error {
			if err := docker.CliRmi(ctx, img.Name()); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	img.SetName(newImageName)

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return err
	} else {
		img.SetInfo(info)
	}

	desc := img.GetStageDescription()

	if desc != nil {
		repository, tag := image.ParseRepositoryAndTag(newImageName)
		desc.Info.Repository = repository
		desc.Info.Tag = tag
	}

	return nil
}

func (runtime *DockerServerBackend) RemoveImage(ctx context.Context, img LegacyImageInterface) error {
	return logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Removing image tag %s", img.Name())).DoError(func() error {
		return runtime.Rmi(ctx, img.Name(), RmiOpts{})
	})
}

func (runtime *DockerServerBackend) PullImageFromRegistry(ctx context.Context, img LegacyImageInterface) error {
	if err := img.Pull(ctx); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", img.Name(), err)
	}

	if info, err := runtime.GetImageInfo(ctx, img.Name(), GetImageInfoOpts{}); err != nil {
		return fmt.Errorf("unable to get inspect of image %s: %w", img.Name(), err)
	} else {
		img.SetInfo(info)
	}

	return nil
}

func (runtime *DockerServerBackend) Tag(ctx context.Context, ref, newRef string, opts TagOpts) error {
	return docker.CliTag(ctx, ref, newRef)
}

func (runtime *DockerServerBackend) Push(ctx context.Context, ref string, opts PushOpts) error {
	return docker.CliPushWithRetries(ctx, ref)
}

func (runtime *DockerServerBackend) Pull(ctx context.Context, ref string, opts PullOpts) error {
	if err := docker.CliPull(ctx, ref); err != nil {
		return fmt.Errorf("unable to pull image %s: %w", ref, err)
	}
	return nil
}

func (runtime *DockerServerBackend) Rmi(ctx context.Context, ref string, opts RmiOpts) error {
	return docker.CliRmi(ctx, ref, "--force")
}

func (runtime *DockerServerBackend) PushImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Pushing %s", img.Name())).DoError(func() error {
		return docker.CliPushWithRetries(ctx, img.Name())
	}); err != nil {
		return err
	}

	return nil
}

// PushBuiltImage is only available for DockerServerBackend
func (runtime *DockerServerBackend) PushBuiltImage(ctx context.Context, img LegacyImageInterface) error {
	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Tagging built image by name %s", img.Name())).DoError(func() error {
		if err := img.TagBuiltImage(ctx); err != nil {
			return fmt.Errorf("unable to tag built image by name %s: %w", img.Name(), err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Pushing %s", img.Name())).DoError(func() error {
		return img.Push(ctx)
	}); err != nil {
		return err
	}

	return nil
}

// TagBuiltImageByName is only available for DockerServerBackend
func (runtime *DockerServerBackend) TagImageByName(ctx context.Context, img LegacyImageInterface) error {
	if img.GetBuiltID() != "" {
		if err := img.TagBuiltImage(ctx); err != nil {
			return fmt.Errorf("unable to tag image %s: %w", img.Name(), err)
		}
	} else {
		if err := runtime.RefreshImageObject(ctx, img); err != nil {
			return err
		}
	}

	return nil
}

func (runtime *DockerServerBackend) String() string {
	return "local-docker-server"
}
