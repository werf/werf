package image

import "fmt"

type StageID struct {
	Signature string `json:"signature"`
	UniqueID  string `json:"uniqueID"`
}

func (id StageID) String() string {
	return fmt.Sprintf("signature:%s uniqueID:%s", id.Signature, id.UniqueID)
}

type StageDescription struct {
	StageID *StageID `json:"stageID"`
	Info    *Info    `json:"info"`
}
