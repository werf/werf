package cleanup_report

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_ReportJSON(t *testing.T) {
	report := NewReport("cleanup", true, "registry.mydomain.com/myproject/werf", "registry.mydomain.com/myproject/werf-final")

	report.AddKept(
		Item{Type: ItemTypeStage, Tag: "1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043", Reason: "used in Kubernetes"},
		Item{Type: ItemTypeStage, Tag: "8c4a1f9b2d7e5a3c6b0d9e8f7a2c1b4d3e6f5a8c9b0d1e2f3a4b5c6d-1749390012345", Reason: "git policy"},
		Item{Type: ItemTypeStage, Tag: "d41d8cd98f00b204e9800998ecf8427e5a1b2c3d4e5f6a7b8c9d0e1f-1749455000111", Reason: "built within last 2 hours"},
		Item{Type: ItemTypeStage, Tag: "aa11bb22cc33dd44ee55ff66aa77bb88cc99dd00ee11ff22aa33bb44-1749300000999", Reason: "import source"},
		Item{Type: ItemTypeCustomTag, Tag: "my-custom-tag"},
	)
	report.AddDeleted(
		Item{Type: ItemTypeStage, Tag: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334"},
		Item{Type: ItemTypeFinalStage, Tag: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334"},
		Item{Type: ItemTypeCustomTag, Tag: "review-1234"},
		Item{Type: ItemTypeRejectedStage, Tag: "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777"},
		Item{Type: ItemTypeImageMetadata, ImageName: "backend", StageID: "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334", Commit: "a3f1c92e4b7d8056f1a2b3c4d5e6f7a8b9c0d1e2"},
		Item{Type: ItemTypeManagedImage, ImageName: "frontend"},
	)

	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, report.Save(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.JSONEq(t, `{
  "apiVersion": "v1",
  "command": "cleanup",
  "dryRun": true,
  "repo": "registry.mydomain.com/myproject/werf",
  "finalRepo": "registry.mydomain.com/myproject/werf-final",
  "kept": [
    { "type": "stage",      "tag": "1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043", "reason": "used in Kubernetes" },
    { "type": "stage",      "tag": "8c4a1f9b2d7e5a3c6b0d9e8f7a2c1b4d3e6f5a8c9b0d1e2f3a4b5c6d-1749390012345", "reason": "git policy" },
    { "type": "stage",      "tag": "d41d8cd98f00b204e9800998ecf8427e5a1b2c3d4e5f6a7b8c9d0e1f-1749455000111", "reason": "built within last 2 hours" },
    { "type": "stage",      "tag": "aa11bb22cc33dd44ee55ff66aa77bb88cc99dd00ee11ff22aa33bb44-1749300000999", "reason": "import source" },
    { "type": "customTag",  "tag": "my-custom-tag" }
  ],
  "deleted": [
    { "type": "stage",         "tag": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334" },
    { "type": "finalStage",    "tag": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334" },
    { "type": "customTag",     "tag": "review-1234" },
    { "type": "rejectedStage", "tag": "0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0bad0-1747999888777" },
    { "type": "imageMetadata", "imageName": "backend", "stageID": "ff00112233445566778899aabbccddeeff00112233445566778899aa-1748001122334", "commit": "a3f1c92e4b7d8056f1a2b3c4d5e6f7a8b9c0d1e2" },
    { "type": "managedImage",  "imageName": "frontend" }
  ]
}`, string(data))
}

func TestAI_HostReportJSON(t *testing.T) {
	report := NewHostReport("host cleanup", false)

	report.AddDeleted(Item{Type: ItemTypeVolume, ID: "e4f1c0a9b7d6"})
	report.AddDeleted(Item{Type: ItemTypeImage, ID: "sha256:8f7e6d5c4b3a2918f0e1d2c3b4a59687766554433221100ffeeddccbbaa9988"})
	report.AddDeleted(Item{Type: ItemTypeContainer, ID: "3a9f1c0e7b2d"})
	report.AddSpaceReclaimed(8271948700)
	report.AddSpaceReclaimed(100)

	path := filepath.Join(t.TempDir(), "host-report.json")
	require.NoError(t, report.Save(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.JSONEq(t, `{
  "apiVersion": "v1",
  "command": "host cleanup",
  "dryRun": false,
  "spaceReclaimed": 8271948800,
  "deleted": [
    { "type": "volume",    "id": "e4f1c0a9b7d6" },
    { "type": "image",     "id": "sha256:8f7e6d5c4b3a2918f0e1d2c3b4a59687766554433221100ffeeddccbbaa9988" },
    { "type": "container", "id": "3a9f1c0e7b2d" }
  ]
}`, string(data))
}

func TestAI_ReportEmptyListsAndOmittedFinalRepo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, NewReport("purge", false, "example.com/repo", "").Save(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.JSONEq(t, `{
  "apiVersion": "v1",
  "command": "purge",
  "dryRun": false,
  "repo": "example.com/repo",
  "kept": [],
  "deleted": []
}`, string(data))
}

func TestAI_NilReportIsNoOp(t *testing.T) {
	var report *Report
	var hostReport *HostReport

	path := filepath.Join(t.TempDir(), "report.json")

	assert.NotPanics(t, func() {
		report.AddKept(Item{Type: ItemTypeStage, Tag: "tag"})
		report.AddDeleted(Item{Type: ItemTypeStage, Tag: "tag"})
		require.NoError(t, report.Save(path))

		hostReport.AddDeleted(Item{Type: ItemTypeImage, ID: "id"})
		hostReport.AddSpaceReclaimed(42)
		require.NoError(t, hostReport.Save(path))
	})

	assert.NoFileExists(t, path)
}

func TestAI_ConcurrentAdd(t *testing.T) {
	report := NewReport("cleanup", false, "example.com/repo", "")
	hostReport := NewHostReport("host cleanup", false)

	const workers = 50
	const perWorker = 20

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				report.AddKept(Item{Type: ItemTypeStage, Tag: fmt.Sprintf("kept-%d-%d", w, i)})
				report.AddDeleted(Item{Type: ItemTypeStage, Tag: fmt.Sprintf("deleted-%d-%d", w, i)})
				hostReport.AddDeleted(Item{Type: ItemTypeImage, ID: fmt.Sprintf("image-%d-%d", w, i)})
				hostReport.AddSpaceReclaimed(1)
			}
		}(w)
	}
	wg.Wait()

	assert.Len(t, report.Kept, workers*perWorker)
	assert.Len(t, report.Deleted, workers*perWorker)
	assert.Len(t, hostReport.Deleted, workers*perWorker)
	assert.Equal(t, uint64(workers*perWorker), hostReport.SpaceReclaimed)
}
