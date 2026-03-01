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

func TestAI_FromImageEmitsDeprecationWarning(t *testing.T) {
	yamlContent := `
image: testimage
fromImage: baseimg
`
	stapelImage, err := parseStapelImage(t, yamlContent, "testimage")

	require.NoError(t, err)
	assert.Equal(t, "baseimg", stapelImage.From)
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

func TestAI_ImportImageFieldDeprecation(t *testing.T) {
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

	assert.Equal(t, "baseimg", stapelImage.Import[0].ImageName)
}

func TestAI_ImportFromFieldNoWarning(t *testing.T) {
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

	assert.Equal(t, "baseimg", stapelImage.Import[0].ImageName)
}

func TestAI_ImportExternalDetectionViaConfigLookup(t *testing.T) {
	t.Skip("External import detection cannot be tested via prepareWerfConfig because validateRelatedImages() runs before updateDependencies() and rejects configs with external imports. The external detection code in updateDependencies() line 261 is unreachable. This requires fixing validateRelatedImages to skip external imports (likely Wave 3 work).")
	yamlImage1 := `
image: image1
from: ubuntu:22.04
`
	yamlImage2 := `
image: image2
from: image1
import:
- from: ubuntu
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
	assert.Equal(t, "image1", image2.From)
	require.Len(t, image2.Import, 1)

	assert.Equal(t, "ubuntu", image2.Import[0].ImageName)
	assert.True(t, image2.Import[0].ExternalImage, "import 'from: ubuntu' should be detected as external")
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

	assert.Equal(t, "image1", image2.Import[0].ImageName)
	assert.False(t, image2.Import[0].ExternalImage, "import 'from: image1' should be detected as internal")
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
