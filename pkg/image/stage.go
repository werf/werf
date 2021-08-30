package image

import (
	"fmt"
	"strconv"
	"time"
)

type StageID struct {
	Digest   string `json:"digest"`
	UniqueID int64  `json:"uniqueID"`
}

func (id StageID) String() string {
	return fmt.Sprintf("%s-%d", id.Digest, id.UniqueID)
}

func (id StageID) UniqueIDAsTime() time.Time {
	return time.Unix(id.UniqueID/1000, id.UniqueID%1000)
}

func (id StageID) IsEqual(another StageID) bool {
	return (id.Digest == another.Digest) && (id.UniqueID == another.UniqueID)
}

type StageDescription struct {
	StageID *StageID `json:"stageID"`
	Info    *Info    `json:"info"`
}

func ParseUniqueIDAsTimestamp(uniqueID string) (int64, error) {
	if timestamp, err := strconv.ParseInt(uniqueID, 10, 64); err != nil {
		return 0, err
	} else {
		return timestamp, nil
	}
}
