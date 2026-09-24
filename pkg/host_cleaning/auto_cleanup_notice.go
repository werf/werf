package host_cleaning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type autoCleanupNotice struct {
	RanAt          time.Time `json:"ran_at"`
	BackendName    string    `json:"backend_name"`
	StoragePath    string    `json:"storage_path"`
	ImagesDeleted  int       `json:"images_deleted"`
	SpaceReclaimed uint64    `json:"space_reclaimed"`
	UsedBytes      uint64    `json:"used_bytes"`
	TotalBytes     uint64    `json:"total_bytes"`
	AllowedBytes   uint64    `json:"allowed_bytes"`
}

func autoCleanupNoticeFilename(werfServiceDir string) string {
	return filepath.Join(werfServiceDir, "auto_host_cleanup_notice.json")
}

func newAutoCleanupNotice(backendName string, report RunGCReport) autoCleanupNotice {
	return autoCleanupNotice{
		RanAt:          time.Now(),
		BackendName:    backendName,
		StoragePath:    report.StoragePath,
		ImagesDeleted:  report.ImagesDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
		UsedBytes:      report.UsedBytes,
		TotalBytes:     report.TotalBytes,
		AllowedBytes:   report.AllowedBytes,
	}
}

// isWorthNotifying reports whether the user should learn about this auto cleanup run: either it
// removed something, or it could not bring the storage volume usage down to the allowed level.
func (notice autoCleanupNotice) isWorthNotifying() bool {
	return notice.ImagesDeleted > 0 || notice.SpaceReclaimed > 0 || notice.storageLimitExceeded()
}

func (notice autoCleanupNotice) storageLimitExceeded() bool {
	return notice.UsedBytes > notice.AllowedBytes
}

func (notice autoCleanupNotice) percentage(bytes uint64) float64 {
	if notice.TotalBytes == 0 {
		return 0
	}
	return float64(bytes) / float64(notice.TotalBytes) * 100
}

func (notice autoCleanupNotice) message() string {
	var lines []string

	if notice.ImagesDeleted > 0 {
		lines = append(lines, fmt.Sprintf(
			"Automatic host cleanup ran in background %s: %d local %s images removed, %s freed in total.",
			humanize.Time(notice.RanAt), notice.ImagesDeleted, notice.BackendName, humanize.Bytes(notice.SpaceReclaimed),
		))
		lines = append(lines, fmt.Sprintf(
			"It runs when the volume usage of %s exceeds the allowed level %s (%.2f%%), so recently built images might have been removed and some images might be rebuilt.",
			notice.StoragePath, humanize.Bytes(notice.AllowedBytes), notice.percentage(notice.AllowedBytes),
		))
	} else {
		lines = append(lines, fmt.Sprintf(
			"Automatic host cleanup ran in background %s and freed %s in total.",
			humanize.Time(notice.RanAt), humanize.Bytes(notice.SpaceReclaimed),
		))
		lines = append(lines, fmt.Sprintf(
			"It runs when the volume usage of %s exceeds the allowed level %s (%.2f%%).",
			notice.StoragePath, humanize.Bytes(notice.AllowedBytes), notice.percentage(notice.AllowedBytes),
		))
	}

	if notice.storageLimitExceeded() {
		lines = append(lines, fmt.Sprintf(
			"The volume usage is still above the allowed level after the cleanup: %s (%.2f%%) of %s.",
			humanize.Bytes(notice.UsedBytes), notice.percentage(notice.UsedBytes), humanize.Bytes(notice.TotalBytes),
		))
		lines = append(lines, "Free up disk space on the host, or raise the allowed level with --allowed-backend-storage-volume-usage ($WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE). Until then werf will keep removing images on every run.")
	} else {
		lines = append(lines, fmt.Sprintf(
			"The volume usage is %s (%.2f%%) of %s now. To keep more images, free up disk space on the host or raise the allowed level with --allowed-backend-storage-volume-usage ($WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE).",
			humanize.Bytes(notice.UsedBytes), notice.percentage(notice.UsedBytes), humanize.Bytes(notice.TotalBytes),
		))
	}

	return strings.Join(lines, "\n")
}

func writeAutoCleanupNotice(werfServiceDir, backendName string, report RunGCReport) error {
	notice := newAutoCleanupNotice(backendName, report)
	if !notice.isWorthNotifying() {
		return nil
	}

	data, err := json.Marshal(notice)
	if err != nil {
		return fmt.Errorf("marshal auto host cleanup notice: %w", err)
	}

	filename := autoCleanupNoticeFilename(werfServiceDir)

	// Write atomically, so that a concurrent werf command never reads a half-written notice.
	tmpFile, err := os.CreateTemp(werfServiceDir, filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("create temporary file for %q: %w", filename, err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write %q: %w", tmpFile.Name(), err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close %q: %w", tmpFile.Name(), err)
	}

	if err := os.Chmod(tmpFile.Name(), 0o644); err != nil {
		return fmt.Errorf("chmod %q: %w", tmpFile.Name(), err)
	}

	if err := os.Rename(tmpFile.Name(), filename); err != nil {
		return fmt.Errorf("rename %q to %q: %w", tmpFile.Name(), filename, err)
	}

	return nil
}

// PopAutoCleanupNotice returns a message about the last auto host cleanup run and whether it
// requires the user's attention, then forgets the run, so the user is notified about it only once.
// An empty message means there is nothing to notify about.
func PopAutoCleanupNotice(_ context.Context, werfServiceDir string) (string, bool, error) {
	filename := autoCleanupNoticeFilename(werfServiceDir)

	// Claim the notice by renaming it, so that neither a concurrent reader reports it twice nor a
	// concurrent cleanup loses a notice written between the read and the removal.
	claimedFilename := fmt.Sprintf("%s.%d", filename, os.Getpid())
	if err := os.Rename(filename, claimedFilename); errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("rename %q to %q: %w", filename, claimedFilename, err)
	}
	defer os.Remove(claimedFilename)

	data, err := os.ReadFile(claimedFilename)
	if err != nil {
		return "", false, fmt.Errorf("read %q: %w", claimedFilename, err)
	}

	var notice autoCleanupNotice
	if err := json.Unmarshal(data, &notice); err != nil {
		return "", false, fmt.Errorf("unmarshal %q: %w", claimedFilename, err)
	}

	return notice.message(), notice.storageLimitExceeded(), nil
}
