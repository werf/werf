//go:build ai_tests

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
)

func TestMain(m *testing.M) {
	parentStack = util.NewStack()
	m.Run()
}

func newTestGiterminismManager() giterminism_manager.Interface {
	return NewGiterminismManagerStub(NewLocalGitRepoStub("test-commit-hash"))
}

func parseStapelImage(t *testing.T, yamlContent, imageName string) (*StapelImage, error) {
	t.Helper()

	doc := &doc{Content: []byte(yamlContent)}
	rawStapelImage := &rawStapelImage{doc: doc}

	err := yaml.Unmarshal(doc.Content, rawStapelImage)
	if err != nil {
		return nil, err
	}

	giterminismManager := newTestGiterminismManager()
	stapelImage, err := rawStapelImage.toStapelImageDirective(giterminismManager, imageName)
	if err != nil {
		return nil, err
	}

	return stapelImage, nil
}

func TestSelfReferenceProducesError(t *testing.T) {
	yamlContent := `
image: testimage
from: testimage
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use itself as base image")
	assert.NotContains(t, err.Error(), ":latest")
}

func TestFromFieldWorksNormally(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	assert.Equal(t, "ubuntu:22.04", stapelImage.From)
}

func TestFromImageFieldIsDeprecatedAlias(t *testing.T) {
	yamlContent := `
image: testimage
fromImage: ubuntu:22.04
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	assert.Equal(t, "ubuntu:22.04", stapelImage.From)
}

func TestFromAndFromImageBothSpecifiedIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
fromImage: alpine:3.18
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fromImage")
}

func TestImportImageFieldIsDeprecatedAlias(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
import:
- image: baseimg
  after: install
  add: /src
  to: /app
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	require.Len(t, stapelImage.Import, 1)
	assert.Equal(t, "baseimg", stapelImage.Import[0].From)
}

func TestImportFromAndImageBothSpecifiedIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
import:
- from: baseimg
  image: otherimg
  after: install
  add: /src
  to: /app
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
}

func TestImportFromFieldWorks(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
import:
- from: baseimg
  after: install
  add: /src
  to: /app
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")
	require.NoError(t, err)
	require.Len(t, stapelImage.Import, 1)

	assert.Equal(t, "baseimg", stapelImage.Import[0].From)
}

func TestExternalFromRequiresTagOrDigest(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must include a tag")
}

func TestExternalFromWithTagIsValid(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.NoError(t, err)
}

func TestExternalFromWithDigestIsValid(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu@sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.NoError(t, err)
}

func TestInternalFromDoesNotRequireTag(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
`
	yamlImage2 := `
image: image2
from: image1
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	doc2 := &doc{Content: []byte(yamlImage2)}
	rawImage2 := &rawStapelImage{doc: doc2}
	err = yaml.Unmarshal(doc2.Content, rawImage2)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1, rawImage2}, nil, meta)
	require.NoError(t, err)
}

func TestExternalImportFromRequiresTagOrDigest(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
import:
- from: alpine
  after: install
  add: /src
  to: /app
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must include a tag")
}

func TestExternalImportFromWithTagIsValid(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
import:
- from: alpine:3.18
  after: install
  add: /src
  to: /app
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.NoError(t, err)
}

func TestImportInternalDetectionViaConfigLookup(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
`
	yamlImage2 := `
image: image2
from: alpine:3.18
import:
- from: image1
  after: install
  add: /src
  to: /app
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	doc2 := &doc{Content: []byte(yamlImage2)}
	rawImage2 := &rawStapelImage{doc: doc2}
	err = yaml.Unmarshal(doc2.Content, rawImage2)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	werfConfig, err := prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1, rawImage2}, nil, meta)
	require.NoError(t, err)

	image2, err := getStapelImageByName(t, werfConfig, "image2")
	require.NoError(t, err)
	assert.Equal(t, "image2", image2.Name)
	require.Len(t, image2.Import, 1)

	assert.Equal(t, "image1", image2.Import[0].From)
	assert.False(t, image2.Import[0].ExternalImage, "import 'from: image1' should be detected as internal")
}

func TestFromScratchIsValid(t *testing.T) {
	yamlImage1 := `
image: image1
from: scratch
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	werfConfig, err := prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.NoError(t, err)

	image1, err := getStapelImageByName(t, werfConfig, "image1")
	require.NoError(t, err)
	assert.Equal(t, "scratch", image1.From)
	assert.False(t, image1.FromExternal, "scratch should not be marked as external")
}

func TestFromScratchWithFromLatestIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
from: scratch
fromLatest: true
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fromLatest")
	assert.Contains(t, err.Error(), "scratch")
}

func TestFromScratchWithFromCacheVersionIsValid(t *testing.T) {
	yamlContent := `
image: testimage
from: scratch
fromCacheVersion: "1"
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	assert.Equal(t, "scratch", stapelImage.From)
	assert.Equal(t, "1", stapelImage.FromCacheVersion)
}

func TestImportFromScratchIsRejected(t *testing.T) {
	yamlImage1 := `
image: image1
from: ubuntu:22.04
import:
- from: scratch
  after: install
  add: /src
  to: /app
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scratch")
}

func TestImageNameScratchIsReserved(t *testing.T) {
	yamlImage1 := `
image: scratch
from: ubuntu:22.04
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestScratchDoesNotRequireTag(t *testing.T) {
	yamlImage1 := `
image: image1
from: scratch
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	err := yaml.Unmarshal(doc1.Content, rawImage1)
	require.NoError(t, err)

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err = prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1}, nil, meta)
	require.NoError(t, err)
}

func parseImageFromDockerfile(t *testing.T, yamlContent, imageName string) (*ImageFromDockerfile, error) {
	t.Helper()

	doc := &doc{Content: []byte(yamlContent)}
	rawDockerfileImage := &rawImageFromDockerfile{doc: doc}

	err := yaml.Unmarshal(doc.Content, rawDockerfileImage)
	if err != nil {
		return nil, err
	}

	giterminismManager := newTestGiterminismManager()
	dockerfileImage, err := rawDockerfileImage.toImageFromDockerfileDirective(giterminismManager, imageName)
	if err != nil {
		return nil, err
	}

	return dockerfileImage, nil
}

func TestDependencyFromFieldWorks(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
dependencies:
- from: baseimg
  before: install
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	require.Len(t, stapelImage.Dependencies, 1)
	assert.Equal(t, "baseimg", stapelImage.Dependencies[0].From)
}

func TestDependencyImageFieldIsDeprecatedAlias(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
dependencies:
- image: baseimg
  before: install
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	require.Len(t, stapelImage.Dependencies, 1)
	assert.Equal(t, "baseimg", stapelImage.Dependencies[0].From)
}

func TestDependencyFromAndImageBothSpecifiedIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
dependencies:
- from: baseimg
  image: otherimg
  before: install
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
}

func TestDockerfileDependencyFromFieldWorks(t *testing.T) {
	yamlContent := `
image: testimage
dockerfile: Dockerfile
dependencies:
- from: baseimg
  imports:
  - type: ImageName
    targetBuildArg: BASE_IMG
`
	dockerfileImage, err := parseImageFromDockerfile(t, yamlContent, "testimage")

	require.NoError(t, err)
	require.Len(t, dockerfileImage.Dependencies, 1)
	assert.Equal(t, "baseimg", dockerfileImage.Dependencies[0].From)
}

func TestDockerfileDependencyImageFieldIsDeprecatedAlias(t *testing.T) {
	yamlContent := `
image: testimage
dockerfile: Dockerfile
dependencies:
- image: baseimg
  imports:
  - type: ImageName
    targetBuildArg: BASE_IMG
`
	dockerfileImage, err := parseImageFromDockerfile(t, yamlContent, "testimage")

	require.NoError(t, err)
	require.Len(t, dockerfileImage.Dependencies, 1)
	assert.Equal(t, "baseimg", dockerfileImage.Dependencies[0].From)
}

func TestDependencyFromPassesPlatformValidator(t *testing.T) {
	yamlImage1 := `
image: base
from: ubuntu:22.04
`
	yamlImage2 := `
image: app
from: ubuntu:22.04
dependencies:
- from: base
  before: install
`

	giterminismManager := newTestGiterminismManager()

	doc1 := &doc{Content: []byte(yamlImage1)}
	rawImage1 := &rawStapelImage{doc: doc1}
	require.NoError(t, yaml.Unmarshal(doc1.Content, rawImage1))

	doc2 := &doc{Content: []byte(yamlImage2)}
	rawImage2 := &rawStapelImage{doc: doc2}
	require.NoError(t, yaml.Unmarshal(doc2.Content, rawImage2))

	meta := &Meta{}
	meta.ConfigVersion = 1
	meta.Project = "test"

	_, err := prepareWerfConfig(giterminismManager, []*rawStapelImage{rawImage1, rawImage2}, nil, meta)
	require.NoError(t, err)
}

func TestDockerfileDependencyFromAndImageBothSpecifiedIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
dockerfile: Dockerfile
dependencies:
- from: baseimg
  image: otherimg
  imports:
  - type: ImageName
    targetBuildArg: BASE_IMG
`
	_, err := parseImageFromDockerfile(t, yamlContent, "testimage")

	require.Error(t, err)
}

func getStapelImageByName(t *testing.T, werfConfig *WerfConfig, name string) (*StapelImage, error) {
	t.Helper()
	for _, img := range werfConfig.GetImageNameList(false) {
		if imageInterface := werfConfig.GetImage(img); imageInterface != nil {
			if stapelImage, ok := imageInterface.(*StapelImage); ok && stapelImage.Name == name {
				return stapelImage, nil
			}
		}
	}
	return nil, assert.AnError
}
