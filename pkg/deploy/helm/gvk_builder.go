package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/phases/stages/externaldeps"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewGVKBuilder(discoveryShortcutExpander meta.RESTMapper) externaldeps.GVKBuilder {
	return &GVKBuilder{
		discoveryShortcutExpander: discoveryShortcutExpander,
	}
}

type GVKBuilder struct {
	discoveryShortcutExpander meta.RESTMapper
}

func (b *GVKBuilder) BuildFromResource(resource string) (*schema.GroupVersionKind, error) {
	gvr, err := b.parseGVR(resource)
	if err != nil {
		return nil, fmt.Errorf("error parsing GroupVersionResource: %w", err)
	}

	gvk, err := b.gvrToGvk(*gvr)
	if err != nil {
		return nil, fmt.Errorf("error converting GroupVersionResource to GroupVersionKind: %w", err)
	}

	return gvk, nil
}

func (b *GVKBuilder) parseGVR(resource string) (*schema.GroupVersionResource, error) {
	var groupVersionResource schema.GroupVersionResource
	if gvr, gr := schema.ParseResourceArg(resource); gvr != nil {
		groupVersionResource = *gvr
	} else {
		if gr.Resource == "" {
			return nil, fmt.Errorf("resource type not specified")
		}

		groupVersionResource = gr.WithVersion("")
	}

	return &groupVersionResource, nil
}

func (b *GVKBuilder) gvrToGvk(groupVersionResource schema.GroupVersionResource) (*schema.GroupVersionKind, error) {
	var groupVersionKind schema.GroupVersionKind
	if preferredKinds, err := b.discoveryShortcutExpander.KindsFor(groupVersionResource); err != nil {
		return nil, fmt.Errorf("error matching a group/version/resource: %w", err)
	} else if len(preferredKinds) == 0 {
		return nil, fmt.Errorf("no matches for group/version/resource")
	} else {
		groupVersionKind = preferredKinds[0]
	}

	return &groupVersionKind, nil
}
