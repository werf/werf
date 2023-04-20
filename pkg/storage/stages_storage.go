package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/pkg/container_backend"
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

type FilterStagesAndProcessRelatedDataOptions struct {
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
}

type StagesStorage interface {
	GetStagesIDs(ctx context.Context, projectName string, opts ...Option) ([]image.StageID, error)
	GetStagesIDsByDigest(ctx context.Context, projectName, digest string, opts ...Option) ([]image.StageID, error)
	GetStageDescription(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDescription, error)
	ExportStage(ctx context.Context, stageDescription *image.StageDescription, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error
	DeleteStage(ctx context.Context, stageDescription *image.StageDescription, options DeleteImageOptions) error

	AddStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error
	CheckStageCustomTag(ctx context.Context, stageDescription *image.StageDescription, tag string) error
	DeleteStageCustomTag(ctx context.Context, tag string) error

	RejectStage(ctx context.Context, projectName, digest string, uniqueID int64) error

	ConstructStageImageName(projectName, digest string, uniqueID int64) string

	// FetchImage will create a local image in the container-runtime
	FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error
	// StoreImage will store a local image into the container-runtime, local built image should exist prior running store
	StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error
	ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error)
	CopyFromStorage(ctx context.Context, src StagesStorage, projectName string, stageID image.StageID, opts CopyFromStorageOptions) (*image.StageDescription, error)

	CreateRepo(ctx context.Context) error
	DeleteRepo(ctx context.Context) error

	// AddManagedImage adds managed image.
	AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	// RmManagedImage removes managed image by imageName or managedImageName.
	// Typically, the managedImageName is the same as the imageName, but may be different
	// if the name contains unsupported special characters, or
	// if the name exceeds the docker tag limit.
	RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	IsManagedImageExist(ctx context.Context, projectName, imageNameOrManagedImageName string, opts ...Option) (bool, error)
	// GetManagedImages returns the list of managedImageName.
	GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error)

	PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error
	RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error
	IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error)
	GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error)

	GetImportMetadata(ctx context.Context, projectName, id string) (*ImportMetadata, error)
	PutImportMetadata(ctx context.Context, projectName string, metadata *ImportMetadata) error
	RmImportMetadata(ctx context.Context, projectName, id string) error
	GetImportMetadataIDs(ctx context.Context, projectName string, opts ...Option) ([]string, error)

	GetClientIDRecords(ctx context.Context, projectName string, opts ...Option) ([]*ClientIDRecord, error)
	PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error
	PostMultiplatformImage(ctx context.Context, projectName, tag string, allPlatformsImages []*image.Info) error
	FilterStagesAndProcessRelatedData(ctx context.Context, stageDescriptions []*image.StageDescription, options FilterStagesAndProcessRelatedDataOptions) ([]*image.StageDescription, error)

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

type CopyFromStorageOptions struct {
	IsMultiplatformImage bool
}
