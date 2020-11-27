package helm

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/werf/logboek"

	"github.com/ghodss/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"helm.sh/helm/v3/pkg/releaseutil"
)

func NewExtraAnnotationsAndLabelsPostRenderer(extraAnnotations, extraLabels map[string]string) *ExtraAnnotationsAndLabelsPostRenderer {
	return &ExtraAnnotationsAndLabelsPostRenderer{
		ExtraAnnotations: extraAnnotations,
		ExtraLabels:      extraLabels,
	}
}

type ExtraAnnotationsAndLabelsPostRenderer struct {
	ExtraAnnotations map[string]string
	ExtraLabels      map[string]string
}

func (pr *ExtraAnnotationsAndLabelsPostRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	splitManifestsByKeys := releaseutil.SplitManifests(renderedManifests.String())

	manifestsKeys := make([]string, 0, len(splitManifestsByKeys))
	for k := range splitManifestsByKeys {
		manifestsKeys = append(manifestsKeys, k)
	}
	sort.Sort(releaseutil.BySplitManifestsOrder(manifestsKeys))

	splitModifiedManifests := make([]string, 0)

	for _, manifestKey := range manifestsKeys {
		manifestContent := splitManifestsByKeys[manifestKey]

		if os.Getenv("WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG") == "1" {
			fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- original manifest BEGIN\n")
			fmt.Printf("%s\n", manifestContent)
			fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- original manifest END\n")
		}

		var obj unstructured.Unstructured

		if err := yaml.Unmarshal([]byte(manifestContent), &obj); err != nil {
			logboek.Warn().LogF("Unable to decode yaml manifest as unstructured object: %s: will not add extra annotations and labels to this object:\n%s\n---\n", err, manifestContent)
			splitModifiedManifests = append(splitModifiedManifests, manifestContent)
			continue
		}

		if obj.GetKind() == "" {
			logboek.Debug().LogF("Skipping emty object\n")
			continue
		}

		if len(pr.ExtraAnnotations) > 0 {
			annotations := obj.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
			}
			for k, v := range pr.ExtraAnnotations {
				annotations[k] = v
			}
			obj.SetAnnotations(annotations)
		}

		if len(pr.ExtraLabels) > 0 {
			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			for k, v := range pr.ExtraLabels {
				labels[k] = v
			}
			obj.SetLabels(labels)
		}

		if modifiedManifestContent, err := yaml.Marshal(obj.Object); err != nil {
			return nil, fmt.Errorf("unable to modify manifest: %s\n%s\n---\n", err, manifestContent)
		} else {
			splitModifiedManifests = append(splitModifiedManifests, string(modifiedManifestContent))

			if os.Getenv("WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG") == "1" {
				fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifest BEGIN\n")
				fmt.Printf("%s\n", modifiedManifestContent)
				fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifest END\n")
			}
		}
	}

	modifiedManifests := bytes.NewBufferString(strings.Join(splitModifiedManifests, "\n---\n"))
	if os.Getenv("WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG") == "1" {
		fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifests RESULT BEGIN\n")
		fmt.Printf("%s\n", modifiedManifests.String())
		fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifests RESULT END\n")
	}

	return modifiedManifests, nil
}

func (pr *ExtraAnnotationsAndLabelsPostRenderer) Add(extraAnnotations, extraLabels map[string]string) {
	if len(extraAnnotations) > 0 {
		if pr.ExtraAnnotations == nil {
			pr.ExtraAnnotations = make(map[string]string)
		}
		for k, v := range extraAnnotations {
			pr.ExtraAnnotations[k] = v
		}
	}

	if len(extraLabels) > 0 {
		if pr.ExtraLabels == nil {
			pr.ExtraLabels = make(map[string]string)
		}
		for k, v := range extraLabels {
			pr.ExtraLabels[k] = v
		}
	}
}
