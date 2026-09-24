package e2e_converge_test

import (
	"context"
	"encoding/json"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/nelm/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type releaseSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
}

type releasesOutput struct {
	Releases []releaseSummary `json:"releases"`
}

type releaseGetOutput struct {
	Release       releaseSummary `json:"release"`
	ResourceSpecs []struct {
		Unstruct struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Data map[string]string `json:"data"`
		} `json:"unstruct"`
	} `json:"resourceSpecs"`
}

func clusterConfigMapData(ctx context.Context, clientFactory *kube.ClientFactory, werfProject *werf.Project) map[string]string {
	ginkgo.GinkgoHelper()

	cm, err := clientFactory.Static().CoreV1().ConfigMaps(werfProject.Namespace(ctx)).Get(ctx, "test1", metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return cm.Data
}

func releaseRevisions(ctx context.Context, werfProject *werf.Project, args []string) []releaseSummary {
	ginkgo.GinkgoHelper()

	var history releasesOutput
	gomega.Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, args, werf.CommonOptions{})), &history)).To(gomega.Succeed())

	return history.Releases
}

func releaseConfigMapData(ctx context.Context, werfProject *werf.Project, args []string) map[string]string {
	ginkgo.GinkgoHelper()

	var get releaseGetOutput
	gomega.Expect(json.Unmarshal([]byte(werfProject.RunCommand(ctx, args, werf.CommonOptions{})), &get)).To(gomega.Succeed())

	for _, spec := range get.ResourceSpecs {
		if spec.Unstruct.Kind == "ConfigMap" && spec.Unstruct.Metadata.Name == "test1" {
			return spec.Unstruct.Data
		}
	}

	ginkgo.Fail("configmap test1 not found in release manifests")

	return nil
}
