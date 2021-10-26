package helm

import (
	"bytes"

	"helm.sh/helm/v3/pkg/postrender"
)

type PostRendererChain struct {
	PostRenderers []postrender.PostRenderer
}

func NewPostRendererChain(postRenderers ...postrender.PostRenderer) *PostRendererChain {
	return &PostRendererChain{
		PostRenderers: postRenderers,
	}
}

func (chain *PostRendererChain) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	newManifests := renderedManifests

	for _, pr := range chain.PostRenderers {
		manifests, err := pr.Run(newManifests)
		if err != nil {
			return manifests, err
		}
		newManifests = manifests
	}

	return newManifests, nil
}
