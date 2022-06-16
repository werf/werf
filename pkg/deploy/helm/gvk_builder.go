package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/phases/stages/externaldeps"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/kubectl/pkg/scheme"
)

func NewGVKBuilder(scheme *runtime.Scheme, shortcutExpander meta.RESTMapper) externaldeps.GVKBuilder {
	return &GVKBuilder{
		scheme:           scheme,
		shortcutExpander: shortcutExpander,
	}
}

type GVKBuilder struct {
	scheme           *runtime.Scheme
	shortcutExpander meta.RESTMapper
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

		if gr.Group != "" {
			if !scheme.Scheme.IsGroupRegistered(gr.Group) {
				return nil, fmt.Errorf("resource group %q is not registered", gr.Group)
			}
			groupVersionResource = scheme.Scheme.PrioritizedVersionsForGroup(gr.Group)[0].WithResource(gr.Resource)
		} else {
			groupVersionResource = gr.WithVersion("")
		}
	}

	return &groupVersionResource, nil
}

func (b *GVKBuilder) gvrToGvk(groupVersionResource schema.GroupVersionResource) (*schema.GroupVersionKind, error) {
	var groupVersionKind schema.GroupVersionKind
	if preferredKinds, err := b.shortcutExpander.KindsFor(groupVersionResource); err != nil {
		return nil, fmt.Errorf("error matching a group/version/resource: %w", err)
	} else if len(preferredKinds) == 0 {
		return nil, fmt.Errorf("no matches for group/version/resource")
	} else {
		groupVersionKind = preferredKinds[0]
	}

	return &groupVersionKind, nil
}
