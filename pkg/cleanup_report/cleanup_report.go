package cleanup_report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type ItemType string

const (
	ItemTypeStage               ItemType = "stage"
	ItemTypeFinalStage          ItemType = "finalStage"
	ItemTypeCustomTag           ItemType = "customTag"
	ItemTypeRejectedStage       ItemType = "rejectedStage"
	ItemTypeRejectedStageMarker ItemType = "rejectedStageMarker"
	ItemTypeImageMetadata       ItemType = "imageMetadata"
	ItemTypeManagedImage        ItemType = "managedImage"
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

	Command   string `json:"command"`
	DryRun    bool   `json:"dryRun"`
	Repo      string `json:"repo"`
	FinalRepo string `json:"finalRepo,omitempty"`
	Kept      []Item `json:"kept"`
	Deleted   []Item `json:"deleted"`
}

type NewReportOptions struct {
	FinalRepo string
}

func NewReport(_ context.Context, command string, dryRun bool, repo string, opts NewReportOptions) *Report {
	return &Report{
		Command:   command,
		DryRun:    dryRun,
		Repo:      repo,
		FinalRepo: opts.FinalRepo,
		Kept:      []Item{},
		Deleted:   []Item{},
	}
}

func (r *Report) AddKept(_ context.Context, items ...Item) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.Kept = append(r.Kept, items...)
}

func (r *Report) AddDeleted(_ context.Context, items ...Item) {
	if r == nil {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.Deleted = append(r.Deleted, items...)
}

func (r *Report) Save(ctx context.Context, path string) error {
	if r == nil {
		return nil
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	return writeJSON(ctx, path, r)
}

// CheckWritable fails before any destructive work when the report could not be written
// afterwards, so a cleanup does not delete objects only to lose its record of them.
func CheckWritable(_ context.Context, path string) error {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return fmt.Errorf("cleanup report %q is not writable: path is a directory", path)
	}

	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("cleanup report %q is not writable: %w", path, err)
	}
	defer os.Remove(f.Name())

	if err := f.Close(); err != nil {
		return fmt.Errorf("close %q: %w", f.Name(), err)
	}

	return nil
}

// writeJSON renames a fully written sibling into place, so a reader never observes a
// truncated report and a failure never touches what is already at the destination.
func writeJSON(_ context.Context, path string, report any) error {
	data, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal cleanup report: %w", err)
	}
	data = append(data, '\n')

	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("create temporary file for cleanup report %q: %w", path, err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write cleanup report to %q: %w", f.Name(), err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close cleanup report %q: %w", f.Name(), err)
	}

	if err := os.Chmod(f.Name(), 0o644); err != nil {
		return fmt.Errorf("chmod cleanup report %q: %w", f.Name(), err)
	}

	if err := os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("rename cleanup report to %q: %w", path, err)
	}

	return nil
}
