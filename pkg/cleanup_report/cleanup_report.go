package cleanup_report

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const APIVersion = "v1"

type ItemType string

const (
	ItemTypeStage               ItemType = "stage"
	ItemTypeFinalStage          ItemType = "finalStage"
	ItemTypeCustomTag           ItemType = "customTag"
	ItemTypeRejectedStage       ItemType = "rejectedStage"
	ItemTypeRejectedStageMarker ItemType = "rejectedStageMarker"
	ItemTypeImageMetadata       ItemType = "imageMetadata"
	ItemTypeManagedImage        ItemType = "managedImage"

	ItemTypeImage     ItemType = "image"
	ItemTypeContainer ItemType = "container"
	ItemTypeVolume    ItemType = "volume"
)

type Item struct {
	Type      ItemType `json:"type"`
	Tag       string   `json:"tag,omitempty"`
	Reason    string   `json:"reason,omitempty"`
	ImageName string   `json:"imageName,omitempty"`
	StageID   string   `json:"stageID,omitempty"`
	Commit    string   `json:"commit,omitempty"`
	ID        string   `json:"id,omitempty"`
}

type Report struct {
	mux sync.Mutex

	APIVersion string `json:"apiVersion"`
	Command    string `json:"command"`
	DryRun     bool   `json:"dryRun"`
	Repo       string `json:"repo"`
	FinalRepo  string `json:"finalRepo,omitempty"`
	Kept       []Item `json:"kept"`
	Deleted    []Item `json:"deleted"`
}

func NewReport(command string, dryRun bool, repo, finalRepo string) *Report {
	return &Report{
		APIVersion: APIVersion,
		Command:    command,
		DryRun:     dryRun,
		Repo:       repo,
		FinalRepo:  finalRepo,
		Kept:       []Item{},
		Deleted:    []Item{},
	}
}

func (r *Report) AddKept(items ...Item) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.Kept = append(r.Kept, items...)
}

func (r *Report) AddDeleted(items ...Item) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.Deleted = append(r.Deleted, items...)
}

func (r *Report) Save(path string) error {
	if r == nil {
		return nil
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	return writeJSON(path, r)
}

type HostReport struct {
	mux sync.Mutex

	APIVersion     string `json:"apiVersion"`
	Command        string `json:"command"`
	DryRun         bool   `json:"dryRun"`
	SpaceReclaimed uint64 `json:"spaceReclaimed"`
	Deleted        []Item `json:"deleted"`
}

func NewHostReport(command string, dryRun bool) *HostReport {
	return &HostReport{
		APIVersion: APIVersion,
		Command:    command,
		DryRun:     dryRun,
		Deleted:    []Item{},
	}
}

func (r *HostReport) AddDeleted(items ...Item) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.Deleted = append(r.Deleted, items...)
}

func (r *HostReport) AddSpaceReclaimed(bytes uint64) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.SpaceReclaimed += bytes
}

func (r *HostReport) Save(path string) error {
	if r == nil {
		return nil
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	return writeJSON(path, r)
}

func writeJSON(path string, report any) error {
	data, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal cleanup report: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cleanup report to %q: %w", path, err)
	}

	return nil
}
