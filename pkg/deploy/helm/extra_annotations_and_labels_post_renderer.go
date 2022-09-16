package helm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	yaml_v3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/werf"
	"github.com/werf/werf/pkg/werf/global_warnings"
)

var WerfRuntimeAnnotations = map[string]string{
	"werf.io/version": werf.Version,
}

var WerfRuntimeLabels = map[string]string{}

func NewExtraAnnotationsAndLabelsPostRenderer(extraAnnotations, extraLabels map[string]string, ignoreInvalidAnnotationsAndLabels bool) *ExtraAnnotationsAndLabelsPostRenderer {
	return &ExtraAnnotationsAndLabelsPostRenderer{
		ExtraAnnotations:                  extraAnnotations,
		ExtraLabels:                       extraLabels,
		IgnoreInvalidAnnotationsAndLabels: ignoreInvalidAnnotationsAndLabels,
		globalWarnings:                    &defaultGlobalWarnings{},
	}
}

type ExtraAnnotationsAndLabelsPostRenderer struct {
	ExtraAnnotations                  map[string]string
	ExtraLabels                       map[string]string
	IgnoreInvalidAnnotationsAndLabels bool

	globalWarnings globalWarnings
}

type defaultGlobalWarnings struct{}

func (gw *defaultGlobalWarnings) GlobalWarningLn(ctx context.Context, msg string) {
	global_warnings.GlobalWarningLn(ctx, msg)
}

type globalWarnings interface {
	GlobalWarningLn(ctx context.Context, msg string)
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

func dereferenceAliasNode(node *yaml_v3.Node) (*yaml_v3.Node, error) {
	dereferencedNode, err := createNode(map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	dereferencedNode.Content = append(dereferencedNode.Content, node.Alias.Content...)

	return dereferencedNode, nil
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

func createNode(v interface{}) (*yaml_v3.Node, error) {
	newNode := &yaml_v3.Node{}
	if err := newNode.Encode(v); err != nil {
		return nil, fmt.Errorf("unable to encode value %#v into yaml node: %w", v, err)
	}
	return newNode, nil
}

func appendToNode(node *yaml_v3.Node, data interface{}) (*yaml_v3.Node, error) {
	newNode, err := createNode(data)
	if err != nil {
		return nil, err
	}

	node.Content = append(node.Content, newNode.Content...)
	return node, nil
}

func validateStringNode(node *yaml_v3.Node) error {
	var v interface{}
	if err := node.Decode(&v); err != nil {
		return fmt.Errorf("unable to decode value %q: %w", node.Value, err)
	}
	if _, ok := v.(string); !ok {
		typeOf := reflect.TypeOf(v)
		if typeOf != nil {
			return fmt.Errorf("invalid node %q: expected string, got %s", node.Value, reflect.TypeOf(v).String())
		} else {
			return fmt.Errorf("invalid node %q: expected string, got null value", node.Value)
		}
	}
	return nil
}

func validateKeyAndValue(keyNode, valueNode *yaml_v3.Node) (errors []error) {
	keyErr := validateStringNode(keyNode)
	if keyErr != nil {
		errors = append(errors, keyErr)
	}

	var valueErr error
	if valueNode.Kind == yaml_v3.AliasNode {
		valueErr = validateStringNode(valueNode.Alias)
	} else {
		valueErr = validateStringNode(valueNode)
	}
	if valueErr != nil {
		errors = append(errors, valueErr)
	}

	return
}

func validateMapStringStringNode(node *yaml_v3.Node) ([]*yaml_v3.Node, []error) {
	content := node.Content
	end := len(content)

	var validValues []*yaml_v3.Node
	var errors []error

	for pos := 0; pos < end; pos += 2 {
		keyNode := content[pos]
		valueNode := content[pos+1]

		if keyNode.Tag == "!!merge" && valueNode.Kind == yaml_v3.AliasNode {
			for pos := 0; pos < len(valueNode.Alias.Content); pos += 2 {
				keyNode := valueNode.Alias.Content[pos]
				valueNode := valueNode.Alias.Content[pos+1]

				newErrs := validateKeyAndValue(keyNode, valueNode)
				if len(newErrs) > 0 {
					errors = append(errors, newErrs...)
				} else {
					validValues = append(validValues, keyNode, valueNode)
				}
			}

			continue
		}

		newErrs := validateKeyAndValue(keyNode, valueNode)
		if len(newErrs) > 0 {
			errors = append(errors, newErrs...)
		} else {
			validValues = append(validValues, keyNode, valueNode)
		}
	}

	return validValues, errors
}

func appendExtraData(node *yaml_v3.Node, key string, data interface{}) error {
	if targetNode := findNodeByKey(node, key); targetNode != nil {
		if targetNode.Kind == yaml_v3.AliasNode {
			dereferencedTargetNode, err := dereferenceAliasNode(targetNode)
			if err != nil {
				return err
			}

			appendToNode(dereferencedTargetNode, data)
			replaceNodeByKey(node, key, dereferencedTargetNode)
		} else {
			appendToNode(targetNode, data)
		}

		if targetNode.Kind != yaml_v3.AliasNode && targetNode.Kind != yaml_v3.MappingNode {
			targetNode.Kind = yaml_v3.MappingNode
		}
		targetNode.Tag = "!!map"
	} else {
		appendToNode(node, map[string]interface{}{key: data})
	}

	return nil
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

		if objMapNode := getMapNode(&objNode); objMapNode != nil {
			if metadataNode := findNodeByKey(objMapNode, "metadata"); metadataNode != nil {

				if obj.IsList() && len(extraAnnotations) > 0 {
					pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("werf annotations won't be applied to *List resource Kinds, including %s. We advise to replace *List resources with multiple separate resources of the same Kind", obj.GetKind()))
				} else if len(extraAnnotations) > 0 {
					if err := appendExtraData(metadataNode, "annotations", extraAnnotations); err != nil {
						pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("werf annotations won't be applied to the %s/%s: an error have occurred during annotations injection: %s\n", strings.ToLower(obj.GetKind()), obj.GetName(), err))
					}
				}

				if obj.IsList() && len(extraLabels) > 0 {
					pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("werf labels won't be applied to *List resource Kinds, including %s. We advise to replace *List resources with multiple separate resources of the same Kind", obj.GetKind()))
				} else if len(extraLabels) > 0 {
					if err := appendExtraData(metadataNode, "labels", extraLabels); err != nil {
						pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("werf labels won't be applied to the %s/%s: an error have occurred during labels injection: %s\n", strings.ToLower(obj.GetKind()), obj.GetName(), err))
					}
				}

				if annotationsNode := findNodeByKey(metadataNode, "annotations"); annotationsNode != nil {
					validNodes, errors := validateMapStringStringNode(annotationsNode)
					for _, err := range errors {
						pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("%s/%s annotations validation: %s", strings.ToLower(obj.GetKind()), obj.GetName(), err.Error()))
					}
					if pr.IgnoreInvalidAnnotationsAndLabels {
						annotationsNode.Content = validNodes
					}
				}

				if labelsNode := findNodeByKey(metadataNode, "labels"); labelsNode != nil {
					validNodes, errors := validateMapStringStringNode(labelsNode)
					for _, err := range errors {
						pr.globalWarnings.GlobalWarningLn(context.Background(), fmt.Sprintf("%s/%s labels validation: %s\n", strings.ToLower(obj.GetKind()), obj.GetName(), err.Error()))
					}
					if pr.IgnoreInvalidAnnotationsAndLabels {
						labelsNode.Content = validNodes
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
				fmt.Printf("%s", modifiedManifestContent.String())
				fmt.Printf("ExtraAnnotationsAndLabelsPostRenderer -- modified manifest END\n")
			}
		}
	}

	modifiedManifests := bytes.NewBufferString(strings.Join(splitModifiedManifests, "---\n"))
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
