package image

import (
	"fmt"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

type StageID struct {
	Digest string `json:"digest"`

	// FIXME: rename to CreationTs / update cacheVersion
	CreationTs      int64 `json:"uniqueID"`
	IsMultiplatform bool  `json:"isMultiplatform"`
}

func NewStageID(digest string, creationTs int64) *StageID {
	return &StageID{
		Digest:          digest,
		CreationTs:      creationTs,
		IsMultiplatform: creationTs == 0,
	}
}

func (id StageID) String() string {
	if id.CreationTs == 0 {
		return id.Digest
	}
	return fmt.Sprintf("%s-%d", id.Digest, id.CreationTs)
}

func (id StageID) CreationTsToTime() time.Time {
	return time.Unix(id.CreationTs/1000, id.CreationTs%1000)
}

func (id StageID) IsEqual(another StageID) bool {
	return (id.Digest == another.Digest) && (id.CreationTs == another.CreationTs)
}

type StageDesc struct {
	StageID *StageID `json:"stageID"`
	Info    *Info    `json:"info"`
}

func ParseCreationTs(creationTs string) (int64, error) {
	if timestamp, err := strconv.ParseInt(creationTs, 10, 64); err != nil {
		return 0, err
	} else {
		return timestamp, nil
	}
}

func (desc *StageDesc) GetCopy() *StageDesc {
	return &StageDesc{
		StageID: NewStageID(desc.StageID.Digest, desc.StageID.CreationTs),
		Info:    desc.Info.GetCopy(),
	}
}

type StageDescSet mapset.Set[*StageDesc]

func NewStageDescSet(descList ...*StageDesc) StageDescSet {
	return mapset.NewSet[*StageDesc](descList...)
}
