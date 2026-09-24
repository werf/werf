package cleaning

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage/manager"
)

var _ manager.StorageManagerInterface = (*fakeStorageManager)(nil)

func newTestStageDesc(repository string, stageID *image.StageID) *image.StageDesc {
	return &image.StageDesc{
		StageID: stageID,
		Info: &image.Info{
			Repository: repository,
			Tag:        stageID.String(),
			Name:       repository + ":" + stageID.String(),
		},
	}
}

func newFakeContextClient(contextName, contextNamespace string, podImagesByNamespace map[string][]string) *ContextClient {
	var objects []runtime.Object
	for namespace, images := range podImagesByNamespace {
		for ind, img := range images {
			objects = append(objects, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-pod-%d", namespace, ind), Namespace: namespace},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "app", Image: img}},
				},
			})
		}
	}

	return &ContextClient{
		ContextName:      contextName,
		ContextNamespace: contextNamespace,
		Client:           fake.NewClientset(objects...),
	}
}

func failPodsListInNamespace(contextClient *ContextClient, namespace string) {
	contextClient.Client.(*fake.Clientset).PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() != namespace {
			return false, nil, nil
		}

		return true, nil, fmt.Errorf("pods is forbidden in namespace %q", namespace)
	})
}

func (f *fakeStorageManager) ForEachDeleteFinalStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedFinalStages = append(f.deletedFinalStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}
