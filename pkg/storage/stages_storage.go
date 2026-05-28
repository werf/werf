package storage

import (
	"errors"
	"fmt"
)

const (
	LocalStorageAddress             = ":local"
	DefaultKubernetesStorageAddress = "kubernetes://werf-synchronization"
	NamelessImageRecordTag          = "__nameless__"
)

var (
	ErrBrokenImage               = errors.New("broken image")
	ErrStageNotFound             = errors.New("stage not found")
	ErrStageRejected             = errors.New("stage rejected")
	ErrImportMetadataNotFound    = errors.New("import metadata not found")
	ErrCustomTagMetadataNotFound = errors.New("custom tag metadata not found")
)

func IsErrBrokenImage(err error) bool {
	return errors.Is(err, ErrBrokenImage)
}

func IsErrStageNotFound(err error) bool {
	return errors.Is(err, ErrStageNotFound)
}

func IsErrStageUnavailable(err error) bool {
	return errors.Is(err, ErrStageNotFound) || errors.Is(err, ErrBrokenImage) || errors.Is(err, ErrStageRejected)
}

func IsErrImportMetadataNotFound(err error) bool {
	return errors.Is(err, ErrImportMetadataNotFound)
}

func IsErrCustomTagMetadataNotFound(err error) bool {
	return errors.Is(err, ErrCustomTagMetadataNotFound)
}

type FilterStagesAndProcessRelatedDataOptions struct {
	SkipUsedImage            bool
	RmForce                  bool
	RmContainersThatUseImage bool
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

type SyncServerRecord struct {
	Server            string
	TimestampMillisec int64
}

type CleanupRecord struct {
	TimestampMillisec int64
}
