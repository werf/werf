package build

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAI_EnvBuildReport_RoundTripNamedImages(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("frontend", newTestReportImageRecord("frontend", true))
	report.SetImageRecord("backend", newTestReportImageRecord("backend", true))

	envData := report.ToEnvFileData()
	parsed, err := parseEnvFileBuildReport(bytes.NewReader(envData))
	require.NoError(t, err)
	require.NotNil(t, parsed)
	require.Len(t, parsed.Images, 2)

	frontendRecord, ok := parsed.Images["frontend"]
	require.True(t, ok)
	assertReportImageRecordEqual(t, report.Images["frontend"], frontendRecord)

	backendRecord, ok := parsed.Images["backend"]
	require.True(t, ok)
	assertReportImageRecordEqual(t, report.Images["backend"], backendRecord)
}

func TestAI_EnvBuildReport_RoundTripUnnamedImage(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("", newTestReportImageRecord("", true))

	envData := report.ToEnvFileData()
	assert.Contains(t, string(envData), "WERF_DOCKER_IMAGE_NAME=")

	parsed, err := parseEnvFileBuildReport(bytes.NewReader(envData))
	require.NoError(t, err)
	require.NotNil(t, parsed)
	require.Len(t, parsed.Images, 1)

	unnamedRecord, ok := parsed.Images[""]
	require.True(t, ok)
	assertReportImageRecordEqual(t, report.Images[""], unnamedRecord)
}

func TestAI_EnvBuildReport_PreservesFinalField(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("frontend", newTestReportImageRecord("frontend", true))
	report.SetImageRecord("backend", newTestReportImageRecord("backend", false))

	envData := report.ToEnvFileData()
	parsed, err := parseEnvFileBuildReport(bytes.NewReader(envData))
	require.NoError(t, err)
	require.NotNil(t, parsed)
	require.Len(t, parsed.Images, 2)

	assert.Equal(t, true, parsed.Images["frontend"].Final)
	assert.Equal(t, false, parsed.Images["backend"].Final)
}

func TestAI_EnvBuildReport_LoadBuildReportFromFile_EnvExtension(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("frontend", newTestReportImageRecord("frontend", true))

	envData := report.ToEnvFileData()
	reportPath := filepath.Join(t.TempDir(), "report.env")
	require.NoError(t, os.WriteFile(reportPath, envData, 0o644))

	loadedReport, err := LoadBuildReportFromFile(context.Background(), reportPath)
	require.NoError(t, err)
	require.NotNil(t, loadedReport)
}

func TestAI_EnvBuildReport_LoadBuildReportFromFile_JSONExtension(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("frontend", newTestReportImageRecord("frontend", true))

	jsonData, err := report.ToJsonData()
	require.NoError(t, err)

	reportPath := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, os.WriteFile(reportPath, jsonData, 0o644))

	loadedReport, err := LoadBuildReportFromFile(context.Background(), reportPath)
	require.NoError(t, err)
	require.NotNil(t, loadedReport)
}

func TestAI_EnvBuildReport_ValidateParsedEnv(t *testing.T) {
	report := NewImagesReport()
	report.SetImageRecord("frontend", newTestReportImageRecord("frontend", true))

	envData := report.ToEnvFileData()
	parsed, err := parseEnvFileBuildReport(bytes.NewReader(envData))
	require.NoError(t, err)
	require.NotNil(t, parsed)

	require.NoError(t, validateBuildReport(context.Background(), "test.env", parsed))
}

func newTestReportImageRecord(werfImageName string, final bool) ReportImageRecord {
	suffix := werfImageName
	if suffix == "" {
		suffix = "unnamed"
	}

	return ReportImageRecord{
		WerfImageName:     werfImageName,
		DockerRepo:        "registry.example.com/" + suffix,
		DockerTag:         "v1",
		DockerImageID:     "sha256:id-" + suffix,
		DockerImageDigest: "sha256:digest-" + suffix,
		DockerImageName:   "registry.example.com/" + suffix + ":v1",
		Final:             final,
	}
}

func assertReportImageRecordEqual(t *testing.T, expected, actual ReportImageRecord) {
	t.Helper()

	assert.Equal(t, expected.WerfImageName, actual.WerfImageName)
	assert.Equal(t, expected.ConfigType, actual.ConfigType)
	assert.Equal(t, expected.DockerRepo, actual.DockerRepo)
	assert.Equal(t, expected.DockerTag, actual.DockerTag)
	assert.Equal(t, expected.DockerImageID, actual.DockerImageID)
	assert.Equal(t, expected.DockerImageDigest, actual.DockerImageDigest)
	assert.Equal(t, expected.DockerImageName, actual.DockerImageName)
	assert.Equal(t, expected.TargetPlatform, actual.TargetPlatform)
	assert.Equal(t, expected.Rebuilt, actual.Rebuilt)
	assert.Equal(t, expected.Final, actual.Final)
	assert.Equal(t, expected.Size, actual.Size)
	assert.Equal(t, expected.BuildTime, actual.BuildTime)
	assert.Equal(t, expected.Stages, actual.Stages)
}
