/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package registry // import "helm.sh/helm/v3/internal/experimental/registry"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	orascontent "github.com/deislabs/oras/pkg/content"
	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"github.com/werf/3p-helm/pkg/chart"
	"github.com/werf/3p-helm/pkg/chart/loader"
	"github.com/werf/3p-helm/pkg/chartutil"
	"github.com/werf/3p-helm/pkg/werf/helmopts"
	"github.com/werf/werf/v2/pkg/ref"
)

const (
	// CacheRootDir is the root directory for a cache
	CacheRootDir = "cache"
)

type (
	// Cache handles local/in-memory storage of Helm charts, compliant with OCI Layout
	Cache struct {
		debug       bool
		out         io.Writer
		rootDir     string
		ociStore    *orascontent.OCIStore
		memoryStore *orascontent.Memorystore
	}

	// CacheRefSummary contains as much info as available describing a chart reference in cache
	// Note: fields here are sorted by the order in which they are set in FetchReference method
	CacheRefSummary struct {
		Name         string
		Repo         string
		Tag          string
		Exists       bool
		Manifest     *ocispec.Descriptor
		Config       *ocispec.Descriptor
		ContentLayer *ocispec.Descriptor
		Size         int64
		Digest       digest.Digest
		CreatedAt    time.Time
		Chart        *chart.Chart
	}
)

// NewCache returns a new OCI Layout-compliant cache with config
func NewCache(opts ...CacheOption) (*Cache, error) {
	cache := &Cache{
		out: ioutil.Discard,
	}
	for _, opt := range opts {
		opt(cache)
	}
	// validate
	if cache.rootDir == "" {
		return nil, errors.New("must set cache root dir on initialization")
	}
	return cache, nil
}

// FetchReference retrieves a chart ref from cache
func (cache *Cache) FetchReference(ctx context.Context, ref *ref.Reference, opts helmopts.HelmOptions) (*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	r := CacheRefSummary{
		Name: ref.FullName(),
		Repo: ref.Repo,
		Tag:  ref.Tag,
	}
	for _, desc := range cache.ociStore.ListReferences() {
		desc := desc

		if desc.Annotations[ocispec.AnnotationRefName] == r.Name {
			r.Exists = true
			manifestBytes, err := cache.fetchBlob(ctx, &desc)
			if err != nil {
				return &r, err
			}
			var manifest ocispec.Manifest
			err = json.Unmarshal(manifestBytes, &manifest)
			if err != nil {
				return &r, err
			}
			r.Manifest = &desc
			r.Config = &manifest.Config
			numLayers := len(manifest.Layers)
			if numLayers != 1 {
				return &r, errors.New(
					fmt.Sprintf("manifest does not contain exactly 1 layer (total: %d)", numLayers))
			}
			var contentLayer *ocispec.Descriptor
			for _, layer := range manifest.Layers {
				layer := layer
				if layer.MediaType == HelmChartContentLayerMediaType || layer.MediaType == HelmChartContentLayerFullMediaType {
					contentLayer = &layer
				}
			}
			if contentLayer == nil {
				return &r, errors.New(
					fmt.Sprintf("manifest does not contain a layer with mediatype %s", HelmChartContentLayerMediaType))
			}
			if contentLayer.Size == 0 {
				return &r, errors.New(
					fmt.Sprintf("manifest layer with mediatype %s is of size 0", HelmChartContentLayerMediaType))
			}
			r.ContentLayer = contentLayer
			info, err := cache.ociStore.Info(withOrasLogger(ctx, cache.out, cache.debug), contentLayer.Digest)
			if err != nil {
				return &r, err
			}
			r.Size = info.Size
			r.Digest = info.Digest
			r.CreatedAt = info.CreatedAt
			contentBytes, err := cache.fetchBlob(ctx, contentLayer)
			if err != nil {
				return &r, err
			}
			ch, err := loader.LoadArchive(bytes.NewBuffer(contentBytes), opts)
			if err != nil {
				return &r, err
			}
			r.Chart = ch
		}
	}
	return &r, nil
}

// StoreReference stores a chart ref in cache
func (cache *Cache) StoreReference(ctx context.Context, ref *ref.Reference, ch *chart.Chart, opts helmopts.HelmOptions) (*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	r := CacheRefSummary{
		Name:  ref.FullName(),
		Repo:  ref.Repo,
		Tag:   ref.Tag,
		Chart: ch,
	}
	existing, _ := cache.FetchReference(ctx, ref, opts)
	r.Exists = existing.Exists
	config, _, err := cache.saveChartConfig(ctx, ch)
	if err != nil {
		return &r, err
	}
	r.Config = config
	contentLayer, _, err := cache.saveChartContentLayer(ctx, ch)
	if err != nil {
		return &r, err
	}
	r.ContentLayer = contentLayer
	info, err := cache.ociStore.Info(withOrasLogger(ctx, cache.out, cache.debug), contentLayer.Digest)
	if err != nil {
		return &r, err
	}
	r.Size = info.Size
	r.Digest = info.Digest
	r.CreatedAt = info.CreatedAt
	manifest, _, err := cache.saveChartManifest(ctx, config, contentLayer)
	if err != nil {
		return &r, err
	}
	r.Manifest = manifest
	return &r, nil
}

// DeleteReference deletes a chart ref from cache
// TODO: garbage collection, only manifest removed
func (cache *Cache) DeleteReference(ctx context.Context, ref *ref.Reference, opts helmopts.HelmOptions) (*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	r, err := cache.FetchReference(ctx, ref, opts)
	if err != nil || !r.Exists {
		return r, err
	}
	cache.ociStore.DeleteReference(r.Name)
	err = cache.ociStore.SaveIndex()
	return r, err
}

// ListReferences lists all chart refs in a cache
func (cache *Cache) ListReferences(ctx context.Context, opts helmopts.HelmOptions) ([]*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	var rr []*CacheRefSummary
	for _, desc := range cache.ociStore.ListReferences() {
		name := desc.Annotations[ocispec.AnnotationRefName]
		if name == "" {
			if cache.debug {
				fmt.Fprintf(cache.out, "warning: found manifest without name: %s", desc.Digest.Hex())
			}
			continue
		}
		ref, err := ref.ParseReference(name)
		if err != nil {
			return rr, err
		}
		r, err := cache.FetchReference(ctx, ref, opts)
		if err != nil {
			return rr, err
		}
		rr = append(rr, r)
	}
	return rr, nil
}

// AddManifest provides a manifest to the cache index.json
func (cache *Cache) AddManifest(ctx context.Context, ref *ref.Reference, manifest *ocispec.Descriptor) error {
	if err := cache.init(); err != nil {
		return err
	}
	cache.ociStore.AddReference(ref.FullName(), *manifest)
	err := cache.ociStore.SaveIndex()
	return err
}

// Provider provides a valid containerd Provider
func (cache *Cache) Provider() content.Provider {
	return content.Provider(cache.ociStore)
}

// Ingester provides a valid containerd Ingester
func (cache *Cache) Ingester() content.Ingester {
	return content.Ingester(cache.ociStore)
}

// ProvideIngester provides a valid oras ProvideIngester
func (cache *Cache) ProvideIngester() orascontent.ProvideIngester {
	return orascontent.ProvideIngester(cache.ociStore)
}

// init creates files needed necessary for OCI layout store
func (cache *Cache) init() error {
	if cache.ociStore == nil {
		ociStore, err := orascontent.NewOCIStore(cache.rootDir)
		if err != nil {
			return err
		}
		cache.ociStore = ociStore
		cache.memoryStore = orascontent.NewMemoryStore()
	}
	return nil
}

// saveChartConfig stores the Chart.yaml as json blob and returns a descriptor
func (cache *Cache) saveChartConfig(ctx context.Context, ch *chart.Chart) (*ocispec.Descriptor, bool, error) {
	configBytes, err := json.Marshal(ch.Metadata)
	if err != nil {
		return nil, false, err
	}
	configExists, err := cache.storeBlob(ctx, configBytes)
	if err != nil {
		return nil, configExists, err
	}
	descriptor := cache.memoryStore.Add("", HelmChartConfigMediaType, configBytes)
	return &descriptor, configExists, nil
}

// saveChartContentLayer stores the chart as tarball blob and returns a descriptor
func (cache *Cache) saveChartContentLayer(ctx context.Context, ch *chart.Chart) (*ocispec.Descriptor, bool, error) {
	destDir := filepath.Join(cache.rootDir, ".build")
	os.MkdirAll(destDir, 0o755)
	tmpFile, err := chartutil.Save(ch, destDir)
	defer os.Remove(tmpFile)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to save")
	}
	contentBytes, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		return nil, false, err
	}
	contentExists, err := cache.storeBlob(ctx, contentBytes)
	if err != nil {
		return nil, contentExists, err
	}
	descriptor := cache.memoryStore.Add("", HelmChartContentLayerMediaType, contentBytes)
	return &descriptor, contentExists, nil
}

// saveChartManifest stores the chart manifest as json blob and returns a descriptor
func (cache *Cache) saveChartManifest(ctx context.Context, config, contentLayer *ocispec.Descriptor) (*ocispec.Descriptor, bool, error) {
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config:    *config,
		Layers:    []ocispec.Descriptor{*contentLayer},
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, false, err
	}
	manifestExists, err := cache.storeBlob(ctx, manifestBytes)
	if err != nil {
		return nil, manifestExists, err
	}
	descriptor := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	return &descriptor, manifestExists, nil
}

// storeBlob stores a blob on filesystem
func (cache *Cache) storeBlob(ctx context.Context, blobBytes []byte) (bool, error) {
	var exists bool
	writer, err := cache.ociStore.Store.Writer(withOrasLogger(ctx, cache.out, cache.debug),
		content.WithRef(digest.FromBytes(blobBytes).Hex()))
	if err != nil {
		return exists, err
	}
	_, err = writer.Write(blobBytes)
	if err != nil {
		return exists, err
	}
	err = writer.Commit(withOrasLogger(ctx, cache.out, cache.debug), 0, writer.Digest())
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return exists, err
		}
		exists = true
	}
	err = writer.Close()
	return exists, err
}

// fetchBlob retrieves a blob from filesystem
func (cache *Cache) fetchBlob(ctx context.Context, desc *ocispec.Descriptor) ([]byte, error) {
	reader, err := cache.ociStore.ReaderAt(withOrasLogger(ctx, cache.out, cache.debug), *desc)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	bytes := make([]byte, desc.Size)
	_, err = reader.ReadAt(bytes, 0)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
