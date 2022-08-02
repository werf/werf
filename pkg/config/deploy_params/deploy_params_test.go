package deploy_params_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/config/deploy_params"
)

var _ = Describe("Deploy params", func() {
	var werfConfig *config.WerfConfig

	BeforeEach(func() {
		werfConfig = &config.WerfConfig{
			Meta: &config.Meta{
				Project: "test-project",
				Deploy: config.MetaDeploy{
					HelmReleaseSlug: newBool(true),
					NamespaceSlug:   newBool(true),
				},
			},
		}
	})

	When("getting helm release", func() {
		When("exact option given", func() {
			It("returns exact option", func() {
				namespace := "some-namespace"
				option := "some-release"
				res, err := deploy_params.GetHelmRelease(option, "", namespace, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(option))
			})

			It("validates invalid option", func() {
				namespace := "some-namespace"
				option := "cannot_use_such_release_name"
				_, err := deploy_params.GetHelmRelease(option, "", namespace, werfConfig)
				Expect(err).To(HaveOccurred())
				Expect(strings.HasPrefix(err.Error(), "bad Helm release specified")).To(BeTrue())
			})
		})

		When("no parameters specified and no custom template", func() {
			It("uses project name as release", func() {
				namespace := "some-namespace"
				res, err := deploy_params.GetHelmRelease("", "", namespace, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(werfConfig.Meta.Project))
			})
		})

		When("environment parameter specified and no custom template", func() {
			It("concatenates project name and environment as release", func() {
				namespace := "namespace"
				environment := "production"
				res, err := deploy_params.GetHelmRelease("", environment, namespace, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(fmt.Sprintf("%s-%s", werfConfig.Meta.Project, environment)))
			})
		})

		When("custom template given", func() {
			It("supports project, environment and namespace go-template params", func() {
				namespace := "some-namespace"
				environment := "production"
				werfConfig.Meta.Deploy.HelmRelease = newString("[[ project ]]-[[ env ]]-[[ namespace ]]")
				res, err := deploy_params.GetHelmRelease("", environment, namespace, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(fmt.Sprintf("%s-%s-%s", werfConfig.Meta.Project, environment, namespace)))
			})
		})
	})

	When("getting namespace", func() {
		When("exact option given", func() {
			It("returns exact option", func() {
				option := "some-namespace"
				res, err := deploy_params.GetKubernetesNamespace(option, "", werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(option))
			})

			It("validates invalid option", func() {
				option := "cannot_use_such_namespace"
				_, err := deploy_params.GetKubernetesNamespace(option, "", werfConfig)
				Expect(err).To(HaveOccurred())
				Expect(strings.HasPrefix(err.Error(), "bad Kubernetes namespace specified")).To(BeTrue())
			})
		})

		When("no parameters specified and no custom template", func() {
			It("uses project name as namespace", func() {
				res, err := deploy_params.GetKubernetesNamespace("", "", werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(werfConfig.Meta.Project))
			})
		})

		When("environment parameter specified and no custom template", func() {
			It("concatenates project name and environment as namespace", func() {
				environment := "production"
				res, err := deploy_params.GetKubernetesNamespace("", environment, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(fmt.Sprintf("%s-%s", werfConfig.Meta.Project, environment)))
			})
		})

		When("custom template given", func() {
			It("supports project, environment and namespace go-template params", func() {
				environment := "production"
				werfConfig.Meta.Deploy.Namespace = newString("[[ env ]]-[[ project ]]")
				res, err := deploy_params.GetKubernetesNamespace("", environment, werfConfig)
				Expect(err).To(Succeed())
				Expect(res).To(Equal(fmt.Sprintf("%s-%s", environment, werfConfig.Meta.Project)))
			})
		})
	})
})

func newBool(v bool) *bool {
	ret := new(bool)
	*ret = v
	return ret
}

func newString(v string) *string {
	ret := new(string)
	*ret = v
	return ret
}
