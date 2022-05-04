package helm

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	yaml_v3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/werf"
)

var WerfRuntimeAnnotations = map[string]string{
	"werf.io/version": werf.Version,
}

var WerfRuntimeLabels = map[string]string{}

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

func replaceNodeByKey(node *yaml_v3.Node, key string, value *yaml_v3.Node) {
	content := node.Content
	end := len(content)

	for pos := 0; pos < end; pos += 2 {
		keyNode := content[pos]

		if keyNode.Tag != "!!str" {
			continue
		}

		var k string
		if err := keyNode.Decode(&k); err != nil {
			continue
		}

		if k == key {
			content[pos+1] = value
			return
		}
	}
}

func findNodeByKey(node *yaml_v3.Node, key string) *yaml_v3.Node {
	content := node.Content
	end := len(content)

	for pos := 0; pos < end; pos += 2 {
		keyNode := content[pos]
		valueNode := content[pos+1]

		if keyNode.Tag != "!!str" {
			continue
		}

		var k string
		if err := keyNode.Decode(&k); err != nil {
			continue
		}

		if k == key {
			return valueNode
		}
	}

	return nil
}

func getMapNode(docNode *yaml_v3.Node) *yaml_v3.Node {
	if docNode.Kind == yaml_v3.DocumentNode {
		if len(docNode.Content) > 0 {
			n := docNode.Content[0]
			if n.Tag == "!!map" {
				return n
			}
		}
	}
	return nil
}

func createNode(v interface{}) *yaml_v3.Node {
	newNode := &yaml_v3.Node{}
	if err := newNode.Encode(v); err != nil {
		logboek.Warn().LogF("Unable to encode map %#v into yaml node: %s\n", v, err)
		return nil
	}
	return newNode
}

func (pr *ExtraAnnotationsAndLabelsPostRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	extraAnnotations := map[string]string{}
	for k, v := range WerfRuntimeAnnotations {
		extraAnnotations[k] = v
	}
	for k, v := range pr.ExtraAnnotations {
		extraAnnotations[k] = v
	}

	extraLabels := map[string]string{}
	for k, v := range WerfRuntimeLabels {
		extraLabels[k] = v
	}
	for k, v := range pr.ExtraLabels {
		extraLabels[k] = v
	}

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
			logboek.Debug().LogF("Skipping empty object\n")
			continue
		}

		var objNode yaml_v3.Node
		if err := yaml_v3.Unmarshal([]byte(manifestContent), &objNode); err != nil {
			logboek.Warn().LogF("Unable to decode yaml manifest as map slice: %s: will not add extra annotations and labels to this object:\n%s\n---\n", err, manifestContent)
			splitModifiedManifests = append(splitModifiedManifests, manifestContent)
			continue
		}

		if os.Getenv("WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG") == "1" {
			fmt.Printf("Unpacket obj annotations: %#v\n", obj.GetAnnotations())
		}

		if obj.IsList() && len(extraAnnotations) > 0 {
			logboek.Warn().LogF("werf annotations won't be applied to *List resource Kinds, including %s. We advise to replace *List resources with multiple separate resources of the same Kind\n", obj.GetKind())
		} else if len(extraAnnotations) > 0 {
			if objMapNode := getMapNode(&objNode); objMapNode != nil {
				if metadataNode := findNodeByKey(objMapNode, "metadata"); metadataNode != nil {
					if annotationsNode := findNodeByKey(metadataNode, "annotations"); annotationsNode != nil {
						if annotationsNode.Kind == yaml_v3.AliasNode {
							if newMetadataAnnotationsNode := createNode(map[string]interface{}{"annotations": map[string]string{}}); newMetadataAnnotationsNode != nil {
								if newAnnotationsNode := findNodeByKey(newMetadataAnnotationsNode, "annotations"); newAnnotationsNode != nil {
									newAnnotationsNode.Content = append(newAnnotationsNode.Content, annotationsNode.Alias.Content...)

									if newExtraAnnotationsNode := createNode(extraAnnotations); newExtraAnnotationsNode != nil {
										newAnnotationsNode.Content = append(newAnnotationsNode.Content, newExtraAnnotationsNode.Content...)
									}
								}
								replaceNodeByKey(metadataNode, "annotations", newMetadataAnnotationsNode)
							}
						} else {
							if newExtraAnnotationsNode := createNode(extraAnnotations); newExtraAnnotationsNode != nil {
								annotationsNode.Content = append(annotationsNode.Content, newExtraAnnotationsNode.Content...)
							}
						}
					} else {
						if newMetadataAnnotationsNode := createNode(map[string]interface{}{"annotations": extraAnnotations}); newMetadataAnnotationsNode != nil {
							metadataNode.Content = append(metadataNode.Content, newMetadataAnnotationsNode.Content...)
						}
					}
				}
			}
		}

		if obj.IsList() && len(extraLabels) > 0 {
			logboek.Warn().LogF("werf labels won't be applied to *List resource Kinds, including %s. We advise to replace *List resources with multiple separate resources of the same Kind\n", obj.GetKind())
		} else if len(extraLabels) > 0 {
			if objMapNode := getMapNode(&objNode); objMapNode != nil {
				if metadataNode := findNodeByKey(objMapNode, "metadata"); metadataNode != nil {
					if labelsNode := findNodeByKey(metadataNode, "labels"); labelsNode != nil {
						if labelsNode.Kind == yaml_v3.AliasNode {
							if newMetadataLabelsNode := createNode(map[string]interface{}{"labels": map[string]string{}}); newMetadataLabelsNode != nil {
								if newLabelsNode := findNodeByKey(newMetadataLabelsNode, "labels"); newLabelsNode != nil {
									newLabelsNode.Content = append(newLabelsNode.Content, labelsNode.Alias.Content...)

									if newExtraLabelsNode := createNode(extraLabels); newExtraLabelsNode != nil {
										newLabelsNode.Content = append(newLabelsNode.Content, newExtraLabelsNode.Content...)
									}
								}
								replaceNodeByKey(metadataNode, "labels", newMetadataLabelsNode)
							}
						} else {
							if newExtraLabelsNode := createNode(extraLabels); newExtraLabelsNode != nil {
								labelsNode.Content = append(labelsNode.Content, newExtraLabelsNode.Content...)
							}
						}
					} else {
						if newMetadataLabelsNode := createNode(map[string]interface{}{"labels": extraLabels}); newMetadataLabelsNode != nil {
							metadataNode.Content = append(metadataNode.Content, newMetadataLabelsNode.Content...)
						}
					}
				}
			}
		}

		var modifiedManifestContent bytes.Buffer
		yamlEncoder := yaml_v3.NewEncoder(&modifiedManifestContent)
		yamlEncoder.SetIndent(2)

		if err := yamlEncoder.Encode(&objNode); err != nil {
			return nil, fmt.Errorf("unable to modify manifest: %w\n%s\n---\n", err, manifestContent)
		} else {
			splitModifiedManifests = append(splitModifiedManifests, modifiedManifestContent.String())

			if os.Getenv("WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG") == "1" {
				fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifest BEGIN\n")
				fmt.Printf("%s\n", modifiedManifestContent.String())
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
