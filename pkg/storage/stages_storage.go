package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
)

const (
	LocalStorageAddress              = ":local"
	DefaultKubernetesStorageAddress  = "kubernetes://werf-synchronization"
	DefaultHttpSynchronizationServer = "https://synchronization.werf.io"
	NamelessImageRecordTag           = "__nameless__"
)

var ErrBrokenImage = errors.New("broken image")

func IsErrBrokenImage(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), ErrBrokenImage.Error())
}

type StagesStorage interface {
	GetStagesIDs(ctx context.Context, projectName string) ([]image.StageID, error)
	GetStagesIDsByDigest(ctx context.Context, projectName, digest string) ([]image.StageID, error)
	GetStageDescription(ctx context.Context, projectName, digest string, uniqueID int64) (*image.StageDescription, error)
	ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string) error
	DeleteStage(ctx context.Context, stageDescription *image.StageDescription, options DeleteImageOptions) error

	AddStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error
	CheckStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error
	DeleteStageCustomTag(ctx context.Context, tag string) error
	GetStageCustomTagMetadataIDs(ctx context.Context) ([]string, error)
	GetStageCustomTagMetadata(ctx context.Context, tagOrID string) (*CustomTagMetadata, error)

	RejectStage(ctx context.Context, projectName, digest string, uniqueID int64) error

	ConstructStageImageName(projectName, digest string, uniqueID int64) string

	// FetchImage will create a local image in the container-runtime
	FetchImage(ctx context.Context, img container_runtime.LegacyImageInterface) error
	// StoreImage will store a local image into the container-runtime, local built image should exist prior running store
	StoreImage(ctx context.Context, img container_runtime.LegacyImageInterface) error
	ShouldFetchImage(ctx context.Context, img container_runtime.LegacyImageInterface) (bool, error)

	CreateRepo(ctx context.Context) error
	DeleteRepo(ctx context.Context) error

	// AddManagedImage adds managed image.
	AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	// RmManagedImage removes managed image by imageName or managedImageName.
	// Typically, the managedImageName is the same as the imageName, but may be different
	// if the name contains unsupported special characters, or
	// if the name exceeds the docker tag limit.
	RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	// GetManagedImages returns the list of managedImageName.
	GetManagedImages(ctx context.Context, projectName string) ([]string, error)

	PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error
	RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error
	IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) (bool, error)
	GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string) (map[string]map[string][]string, map[string]map[string][]string, error)

	GetImportMetadata(ctx context.Context, projectName, id string) (*ImportMetadata, error)
	PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error
	RmImportMetadata(ctx context.Context, projectName, id string) error
	GetImportMetadataIDs(ctx context.Context, projectName string) ([]string, error)

	GetClientIDRecords(ctx context.Context, projectName string) ([]*ClientIDRecord, error)
	PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error

	String() string
	Address() string
}

type ClientIDRecord struct {
	ClientID          string
	TimestampMillisec int64
}

func (rec *ClientIDRecord) String() string {
	return fmt.Sprintf("clientID:%s tsMillisec:%d", rec.ClientID, rec.TimestampMillisec)
}

type ImageMetadata struct {
	ContentDigest string
}

type StagesStorageOptions struct {
	RepoStagesStorageOptions
}

func NewStagesStorage(stagesStorageAddress string, containerRuntime container_runtime.ContainerRuntime, options StagesStorageOptions) (StagesStorage, error) {
	if stagesStorageAddress == LocalStorageAddress {
		return NewDockerServerStagesStorage(containerRuntime.(*container_runtime.DockerServerRuntime)), nil
	} else { // Docker registry based stages storage
		return NewRepoStagesStorage(stagesStorageAddress, containerRuntime, options.RepoStagesStorageOptions)
	}
}
