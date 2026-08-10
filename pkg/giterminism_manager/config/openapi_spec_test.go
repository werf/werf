package config

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFileReader struct {
	exist bool
	data  []byte
}

func (f fakeFileReader) IsGiterminismConfigExistAnywhere(_ context.Context, _ string) (bool, error) {
	return f.exist, nil
}

func (f fakeFileReader) ReadGiterminismConfig(_ context.Context, _ string) ([]byte, error) {
	return f.data, nil
}

func validateConfig(t *testing.T, yamlData string) ([]byte, error) {
	t.Helper()
	data := []byte(yamlData)
	err := processWithOpenAPISchema(&data)
	return data, err
}

func TestValidFullConfig(t *testing.T) {
	data, err := validateConfig(t, `
giterminismConfigVersion: 1
cli:
  allowCustomTags: true
config:
  allowUncommitted: true
  allowUncommittedTemplates:
    - "**/*"
  goTemplateRendering:
    allowEnvVariables:
      - WERF_FOO
    allowUncommittedFiles:
      - "**/*"
  secrets:
    allowEnvVariables:
      - SECRET_ENV
    allowFiles:
      - .secret
    allowValueIds:
      - secret-id
  stapel:
    allowFromLatest: true
    git:
      allowBranch: true
    mount:
      allowBuildDir: true
      allowFromPaths:
        - build
  dockerfile:
    allowUncommitted:
      - Dockerfile
    allowUncommittedDockerignoreFiles:
      - .dockerignore
    allowContextAddFiles:
      - some
helm:
  allowUncommittedFiles:
    - "**/*"
includes:
  allowIncludesUpdate: true
`)
	require.NoError(t, err)

	var c Config
	require.NoError(t, json.Unmarshal(data, &c))
	assert.True(t, c.IsCustomTagsAccepted())
	assert.True(t, c.IsUpdateIncludesAccepted())
	assert.True(t, c.IsConfigSecretEnvAccepted("SECRET_ENV"))
	assert.True(t, c.IsConfigSecretValueAccepted("secret-id"))
}

func TestVersionAsNumberAndString(t *testing.T) {
	_, err := validateConfig(t, "giterminismConfigVersion: 1\n")
	require.NoError(t, err)

	_, err = validateConfig(t, `giterminismConfigVersion: "1"`+"\n")
	require.NoError(t, err)
}

func TestMissingVersionFails(t *testing.T) {
	_, err := validateConfig(t, "cli:\n  allowCustomTags: true\n")
	require.Error(t, err)
}

func TestMinimalConfigUnmarshals(t *testing.T) {
	data, err := validateConfig(t, "giterminismConfigVersion: 1\n")
	require.NoError(t, err)

	var c Config
	require.NoError(t, json.Unmarshal(data, &c))
}

func TestUnknownRootKeyRejected(t *testing.T) {
	_, err := validateConfig(t, `
giterminismConfigVersion: 1
allowCustomTags: true
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allowCustomTags")
}

func TestUnknownNestedKeyRejected(t *testing.T) {
	cases := map[string]string{
		"config": `
giterminismConfigVersion: 1
config:
  allowUncommited: true
`,
		"config.stapel": `
giterminismConfigVersion: 1
config:
  stapel:
    allowFromLatests: true
`,
		"cli": `
giterminismConfigVersion: 1
cli:
  allowCustomTag: true
`,
		"helm": `
giterminismConfigVersion: 1
helm:
  allowUncommittedFile: ["**/*"]
`,
		"config.secrets": `
giterminismConfigVersion: 1
config:
  secrets:
    allowEnvVariable: ["FOO"]
`,
		"includes": `
giterminismConfigVersion: 1
includes:
  allowIncludesUpdat: true
`,
	}

	for name, yamlData := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := validateConfig(t, yamlData)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "forbidden property")
		})
	}
}

func TestConfigSecretsValid(t *testing.T) {
	data, err := validateConfig(t, `
giterminismConfigVersion: 1
config:
  secrets:
    allowEnvVariables:
      - FOO
    allowFiles:
      - .secret
    allowValueIds:
      - vid
`)
	require.NoError(t, err)

	var c Config
	require.NoError(t, json.Unmarshal(data, &c))
	assert.True(t, c.IsConfigSecretEnvAccepted("FOO"))
}

func TestIncludesValid(t *testing.T) {
	data, err := validateConfig(t, `
giterminismConfigVersion: 1
includes:
  allowIncludesUpdate: true
`)
	require.NoError(t, err)

	var c Config
	require.NoError(t, json.Unmarshal(data, &c))
	assert.True(t, c.IsUpdateIncludesAccepted())
}

func TestNoFileReturnsEmptyConfig(t *testing.T) {
	c, err := NewConfig(context.Background(), fakeFileReader{exist: false}, "werf-giterminism.yaml")
	require.NoError(t, err)
	assert.Equal(t, Config{}, c)
}

func TestNewConfigUnknownKeyFails(t *testing.T) {
	_, err := NewConfig(context.Background(), fakeFileReader{
		exist: true,
		data:  []byte("giterminismConfigVersion: 1\nallowCustomTags: true\n"),
	}, "werf-giterminism.yaml")
	require.Error(t, err)
}

func TestWrongTypeRejected(t *testing.T) {
	_, err := validateConfig(t, `
giterminismConfigVersion: 1
cli:
  allowCustomTags: "yes"
`)
	require.Error(t, err)
}
