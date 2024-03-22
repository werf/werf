package docker_registry

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

const (
	AzureCrImplementationName          = "acr"
	azureCrRepositoryNotFoundErrPrefix = "azure cr repository not found: "
)

type AzureCrRepositoryNotFoundErr apiError

func NewAzureCrRepositoryNotFoundErr(err error) AzureCrRepositoryNotFoundErr {
	return AzureCrRepositoryNotFoundErr{
		error: fmt.Errorf(azureCrRepositoryNotFoundErrPrefix + err.Error()),
	}
}

func IsAzureCrRepositoryNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), azureCrRepositoryNotFoundErrPrefix)
}

var azureCrPatterns = []string{`^.*\.azurecr\.io`}

type azureCr struct {
	*defaultImplementation
}

type azureCrOptions struct {
	defaultImplementationOptions
}

func newAzureCr(options azureCrOptions) (*azureCr, error) {
	d, err := newDefaultAPIForImplementation(AzureCrImplementationName, options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	azureCr := &azureCr{defaultImplementation: d}

	return azureCr, nil
}

func (r *azureCr) DeleteRepoImage(ctx context.Context, repoImage *image.Info) error {
	registryName, repository, err := r.parseReference(repoImage.Repository)
	if err != nil {
		return err
	}

	return r.azRun(
		ctx,
		"acr", "repository", "delete",
		"--name="+registryName,
		"--image="+strings.Join([]string{repository, repoImage.Tag}, ":"),
		"--yes",
	)
}

func (r *azureCr) DeleteRepo(ctx context.Context, reference string) error {
	registryName, repository, err := r.parseReference(reference)
	if err != nil {
		return err
	}

	err = r.azRun(
		ctx,
		"acr", "repository", "delete",
		"--name="+registryName,
		"--repository="+repository,
		"--yes",
	)
	if err != nil {
		if strings.Contains(err.Error(), "repository name not known to registry") {
			return NewAzureCrRepositoryNotFoundErr(err)
		}

		return err
	}

	return nil
}

func (r *azureCr) String() string {
	return AzureCrImplementationName
}

func (r *azureCr) parseReference(reference string) (string, string, error) {
	var registryId, repository string

	parsedReference, err := name.NewRepository(reference)
	if err != nil {
		return "", "", err
	}

	if !strings.HasSuffix(parsedReference.RegistryStr(), ".azurecr.io") {
		return "", "", fmt.Errorf("reference %s is not compatible with %s container registry", reference, r.String())
	}

	registryId = strings.TrimSuffix(parsedReference.RegistryStr(), ".azurecr.io")
	repository = parsedReference.RepositoryStr()

	return registryId, repository, nil
}

func (r *azureCr) azRun(ctx context.Context, args ...string) error {
	_, err := exec.LookPath("az")
	if err != nil {
		return err
	}

	command := strings.Join(append([]string{"az"}, args...), " ")
	logboek.Context(ctx).Debug().LogLn(command)
	c := exec.Command("az", args...)

	output, err := c.CombinedOutput()
	logboek.Context(ctx).Debug().LogLn("output:", string(output))

	if err != nil {
		return fmt.Errorf(
			"command: %s\n%s\nerror: %w", command,
			strings.TrimSpace(string(output)),
			err,
		)
	}

	return nil
}
