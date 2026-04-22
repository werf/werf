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

func TestAI_SelfReferenceProducesError(t *testing.T) {
	yamlContent := `
image: testimage
from: testimage
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use itself as base image")
	assert.NotContains(t, err.Error(), ":latest")
}

func TestAI_FromFieldWorksNormally(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	assert.Equal(t, "ubuntu:22.04", stapelImage.From)
}

func TestAI_FromImageFieldIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
fromImage: baseimg
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fromImage")
}

func TestAI_ImportImageFieldIsRejected(t *testing.T) {
	yamlContent := `
image: testimage
from: ubuntu:22.04
import:
- image: baseimg
  after: install
  add: /src
  to: /app
`
	_, err := parseStapelImage(t, yamlContent, "testimage")

	require.Error(t, err)
}

func TestAI_ImportFromFieldWorks(t *testing.T) {
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

func TestAI_ExternalFromRequiresTagOrDigest(t *testing.T) {
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

func TestAI_ExternalFromWithTagIsValid(t *testing.T) {
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

func TestAI_ExternalFromWithDigestIsValid(t *testing.T) {
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

func TestAI_InternalFromDoesNotRequireTag(t *testing.T) {
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

func TestAI_ExternalImportFromRequiresTagOrDigest(t *testing.T) {
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

func TestAI_ExternalImportFromWithTagIsValid(t *testing.T) {
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

func TestAI_ImportInternalDetectionViaConfigLookup(t *testing.T) {
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

func TestAI_FromScratchIsValid(t *testing.T) {
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

func TestAI_FromScratchWithFromLatestIsRejected(t *testing.T) {
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

func TestAI_FromScratchWithFromCacheVersionIsValid(t *testing.T) {
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

func TestAI_ImportFromScratchIsRejected(t *testing.T) {
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

func TestAI_ImageNameScratchIsReserved(t *testing.T) {
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

func TestAI_ScratchDoesNotRequireTag(t *testing.T) {
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
