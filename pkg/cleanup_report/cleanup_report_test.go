package cleanup_report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportJSON(t *testing.T) {
	ctx := context.Background()

	report := NewReport(ctx, "cleanup", true, "registry.mydomain.com/myproject/werf", NewReportOptions{FinalRepo: "registry.mydomain.com/myproject/werf-final"})

	report.AddKept(ctx,
		Item{Type: ItemTypeStage, Tag: "1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043", Reason: "used in Kubernetes"},
		Item{Type: ItemTypeStage, Tag: "aa11bb22cc33dd44ee55ff66aa77bb88cc99dd00ee11ff22aa33bb44-1749300000999", Reason: "import source"},
		Item{Type: ItemTypeFinalStage, Tag: "d41d8cd98f00b204e9800998ecf8427e5a1b2c3d4e5f6a7b8c9d0e1f-1749455000111", Reason: "used in Kubernetes"},
		Item{Type: ItemTypeCustomTag, Tag: "my-custom-tag"},
	)
	report.AddDeleted(ctx,
		Item{Type: ItemTypeStage, Tag: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334"},
		Item{Type: ItemTypeFinalStage, Tag: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334"},
		Item{Type: ItemTypeCustomTag, Tag: "review-1234"},
		Item{Type: ItemTypeRejectedStage, Tag: "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777"},
		Item{Type: ItemTypeRejectedStageMarker, Tag: "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777"},
		Item{Type: ItemTypeImageMetadata, ImageName: "backend", StageID: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334", Commit: "a3f1c92e4b7d8056f1a2b3c4d5e6f7a8b9c0d1e2"},
		Item{Type: ItemTypeManagedImage, ImageName: "frontend"},
	)

	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, report.Save(ctx, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.JSONEq(t, `{
  "command": "cleanup",
  "dryRun": true,
  "repo": "registry.mydomain.com/myproject/werf",
  "finalRepo": "registry.mydomain.com/myproject/werf-final",
  "kept": [
    { "type": "stage",      "tag": "1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043", "reason": "used in Kubernetes" },
    { "type": "stage",      "tag": "aa11bb22cc33dd44ee55ff66aa77bb88cc99dd00ee11ff22aa33bb44-1749300000999", "reason": "import source" },
    { "type": "finalStage", "tag": "d41d8cd98f00b204e9800998ecf8427e5a1b2c3d4e5f6a7b8c9d0e1f-1749455000111", "reason": "used in Kubernetes" },
    { "type": "customTag",  "tag": "my-custom-tag" }
  ],
  "deleted": [
    { "type": "stage",               "tag": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334" },
    { "type": "finalStage",          "tag": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334" },
    { "type": "customTag",           "tag": "review-1234" },
    { "type": "rejectedStage",       "tag": "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777" },
    { "type": "rejectedStageMarker", "tag": "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777" },
    { "type": "imageMetadata",       "imageName": "backend", "stageID": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334", "commit": "a3f1c92e4b7d8056f1a2b3c4d5e6f7a8b9c0d1e2" },
    { "type": "managedImage",        "imageName": "frontend" }
  ]
}`, string(data))
}

func TestReportHasNoEnvelopeOrMetaRepo(t *testing.T) {
	ctx := context.Background()

	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, NewReport(ctx, "cleanup", false, "example.com/app", NewReportOptions{FinalRepo: "example.com/app-final"}).Save(ctx, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	for _, dropped := range []string{"apiVersion", "kind", "metaRepo", "spaceReclaimed", "reference"} {
		assert.NotContains(t, string(data), dropped)
	}
}

func TestReportEmptyListsAndOmittedFinalRepo(t *testing.T) {
	ctx := context.Background()

	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, NewReport(ctx, "purge", false, "example.com/repo", NewReportOptions{}).Save(ctx, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.JSONEq(t, `{
  "command": "purge",
  "dryRun": false,
  "repo": "example.com/repo",
  "kept": [],
  "deleted": []
}`, string(data))
}

func TestNilReportIsNoOp(t *testing.T) {
	ctx := context.Background()

	var report *Report

	path := filepath.Join(t.TempDir(), "report.json")

	assert.NotPanics(t, func() {
		report.AddKept(ctx, Item{Type: ItemTypeStage, Tag: "tag"})
		report.AddDeleted(ctx, Item{Type: ItemTypeStage, Tag: "tag"})
		require.NoError(t, report.Save(ctx, path))
	})

	assert.NoFileExists(t, path)
}

func TestConcurrentAdd(t *testing.T) {
	ctx := context.Background()

	report := NewReport(ctx, "cleanup", false, "example.com/repo", NewReportOptions{})

	const workers = 50
	const perWorker = 20

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				report.AddKept(ctx, Item{Type: ItemTypeStage, Tag: fmt.Sprintf("kept-%d-%d", w, i)})
				report.AddDeleted(ctx, Item{Type: ItemTypeStage, Tag: fmt.Sprintf("deleted-%d-%d", w, i)})
			}
		}(w)
	}
	wg.Wait()

	assert.Len(t, report.Kept, workers*perWorker)
	assert.Len(t, report.Deleted, workers*perWorker)
}

func TestSaveReplacesPreviousReportAtomically(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	require.NoError(t, os.WriteFile(path, []byte("previous\n"), 0o644))

	report := NewReport(ctx, "cleanup", false, "example.com/app", NewReportOptions{})
	report.AddDeleted(ctx, Item{Type: ItemTypeStage, Tag: "tag"})

	require.NoError(t, report.Save(ctx, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "previous")
	assert.Equal(t, byte('\n'), data[len(data)-1])

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "the temporary sibling must not be left behind")
}

func TestSaveFailsLeavingDestinationAndNoTempBehind(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	destination := filepath.Join(dir, "report.json")
	require.NoError(t, os.Mkdir(destination, 0o755))

	sentinel := filepath.Join(destination, "sentinel")
	require.NoError(t, os.WriteFile(sentinel, []byte("untouched\n"), 0o644))

	report := NewReport(ctx, "cleanup", false, "example.com/app", NewReportOptions{})
	report.AddDeleted(ctx, Item{Type: ItemTypeStage, Tag: "tag"})

	err := report.Save(ctx, destination)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rename cleanup report")

	data, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	assert.Equal(t, "untouched\n", string(data), "a failed save must not disturb the destination")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "the temporary sibling must not be left behind")
}

func TestCheckWritable(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	require.NoError(t, CheckWritable(ctx, filepath.Join(dir, "report.json")))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "the probe must not leave a file behind")

	err = CheckWritable(ctx, filepath.Join(dir, "missing-dir", "report.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not writable")
}

func TestCheckWritableRejectsDirectory(t *testing.T) {
	ctx := context.Background()

	destination := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, os.Mkdir(destination, 0o755))

	err := CheckWritable(ctx, destination)
	require.Error(t, err, "a directory at the report path must be rejected before any deletion happens")
	assert.Contains(t, err.Error(), "is not writable")
}
