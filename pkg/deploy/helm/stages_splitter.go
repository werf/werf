package helm

import (
	"fmt"
	"sort"
	"strconv"

	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/phasemanagers/stages"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

type StagesSplitter struct{}

func (s StagesSplitter) Split(resources kube.ResourceList) (stages.SortedStageList, error) {
	var stageList stages.SortedStageList
	if err := resources.Visit(func(res *resource.Info, err error) error {
		if err != nil {
			return err
		}

		unstructuredObj := unstructured.Unstructured{}
		unstructuredObj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(res.Object)
		if err != nil {
			return fmt.Errorf("error converting object to unstructured type: %w", err)
		}

		var weight int
		if w, ok := unstructuredObj.GetAnnotations()[StageWeightAnnoName]; ok {
			weight, err = strconv.Atoi(w)
			if err != nil {
				return fmt.Errorf("error parsing annotation \"%s: %s\" — value should be an integer: %w", StageWeightAnnoName, w, err)
			}
		}

		stage := stageList.StageByWeight(weight)

		if stage == nil {
			stage = &stages.Stage{
				Weight: weight,
			}
			stageList = append(stageList, stage)
		}

		stage.DesiredResources.Append(res)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error visiting resources list: %w", err)
	}

	sort.Sort(stageList)

	return stageList, nil
}
