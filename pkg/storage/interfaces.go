package storage

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/image"
)

type BaseStorage interface {
	String() string
	Address() string
}

type StageReader interface {
	BaseStorage
	ConstructStageImageName(projectName, digest string, creationTs int64) string
	GetStagesIDs(ctx context.Context, projectName string, opts ...Option) ([]image.StageID, error)
	GetStagesIDsByDigest(ctx context.Context, projectName, digest string, parentStageCreationTs int64, opts ...Option) ([]image.StageID, error)
	GetStageDesc(ctx context.Context, projectName string, stageID image.StageID) (*image.StageDesc, error)
	FetchImage(ctx context.Context, img container_backend.LegacyImageInterface) error
	ShouldFetchImage(ctx context.Context, img container_backend.LegacyImageInterface) (bool, error)
}

type StageWriter interface {
	BaseStorage
	ConstructStageImageName(projectName, digest string, creationTs int64) string
	StoreImage(ctx context.Context, img container_backend.LegacyImageInterface) error
	DeleteStage(ctx context.Context, stageDesc *image.StageDesc, options DeleteImageOptions) error
	CopyFromStorage(ctx context.Context, src StageReader, projectName string, stageID image.StageID, opts CopyFromStorageOptions) (*image.StageDesc, error)
	MutateAndPushImage(ctx context.Context, src, dest string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error
	PostManifest(ctx context.Context, ref string, opts container_backend.PostManifestOpts) error
}

type ImagesRepoStorage interface {
	StageReader
	StageWriter
	ExportStage(ctx context.Context, stageDesc *image.StageDesc, destinationReference string, mutateConfigFunc func(config v1.Config) (v1.Config, error)) error
	MutateAndPushImage(ctx context.Context, src, dest string, newConfig image.SpecConfig, stageImage container_backend.LegacyImageInterface) error
	PostManifest(ctx context.Context, ref string, opts container_backend.PostManifestOpts) error
	PostMultiplatformImage(ctx context.Context, projectName, tag string, allPlatformsImages []*image.Info, platforms []string) error
	AddStageCustomTag(ctx context.Context, stageDesc *image.StageDesc, tag string) error
	CheckStageCustomTag(ctx context.Context, stageDesc *image.StageDesc, tag string) error
	DeleteStageCustomTag(ctx context.Context, tag string) error
	CreateRepo(ctx context.Context) error
	DeleteRepo(ctx context.Context) error
	GetStageCustomTagMetadataIDs(ctx context.Context, opts ...Option) ([]string, error)
	GetStageCustomTagMetadata(ctx context.Context, tagOrID string) (*CustomTagMetadata, error)
	RegisterStageCustomTag(ctx context.Context, projectName string, stageDesc *image.StageDesc, tag string) error
	UnregisterStageCustomTag(ctx context.Context, tag string) error
}

type MetaStorage interface {
	BaseStorage
	RejectStage(ctx context.Context, projectName, digest string, creationTs int64) error
	AddManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	RmManagedImage(ctx context.Context, projectName, imageNameOrManagedImageName string) error
	IsManagedImageExist(ctx context.Context, projectName, imageNameOrManagedImageName string, opts ...Option) (bool, error)
	GetManagedImages(ctx context.Context, projectName string, opts ...Option) ([]string, error)
	PutImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string) error
	RmImageMetadata(ctx context.Context, projectName, imageNameOrManagedImageNameOrImageMetadataID, commit, stageID string) error
	IsImageMetadataExist(ctx context.Context, projectName, imageNameOrManagedImageName, commit, stageID string, opts ...Option) (bool, error)
	GetAllAndGroupImageMetadataByImageName(ctx context.Context, projectName string, imageNameOrManagedImageList []string, opts ...Option) (map[string]map[string][]string, map[string]map[string][]string, error)
	GetClientIDRecords(ctx context.Context, projectName string, opts ...Option) ([]*ClientIDRecord, error)
	PostClientIDRecord(ctx context.Context, projectName string, rec *ClientIDRecord) error
	GetSyncServerRecords(ctx context.Context, projectName string, opts ...Option) ([]*SyncServerRecord, error)
	PostSyncServerRecord(ctx context.Context, projectName string, rec *SyncServerRecord) error
	FilterStageDescSetAndProcessRelatedData(ctx context.Context, stageDescSet image.StageDescSet, options FilterStagesAndProcessRelatedDataOptions) (image.StageDescSet, error)
	GetLastCleanupRecord(ctx context.Context, projectName string, opts ...Option) (*CleanupRecord, error)
	PostLastCleanupRecord(ctx context.Context, projectName string) error
}

type CacheAndMetaStorage interface {
	StageReader
	StageWriter
	MetaStorage
}

var (
	_ BaseStorage         = (*RepoStagesStorage)(nil)
	_ StageReader         = (*RepoStagesStorage)(nil)
	_ StageWriter         = (*RepoStagesStorage)(nil)
	_ ImagesRepoStorage   = (*RepoStagesStorage)(nil)
	_ MetaStorage         = (*RepoStagesStorage)(nil)
	_ CacheAndMetaStorage = (*RepoStagesStorage)(nil)

	_ BaseStorage       = (*LocalStagesStorage)(nil)
	_ StageReader       = (*LocalStagesStorage)(nil)
	_ StageWriter       = (*LocalStagesStorage)(nil)
	_ ImagesRepoStorage = (*LocalStagesStorage)(nil)
)
