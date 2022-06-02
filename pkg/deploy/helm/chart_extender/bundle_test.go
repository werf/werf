package chart_extender

import (
	"context"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	helm_v3 "helm.sh/helm/v3/cmd/helm"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/werf/werf/pkg/deploy/secrets_manager"
)

var _ = Describe("Bundle", func() {
	BeforeEach(func() {
		os.Setenv("WERF_SECRET_KEY", "bfd966688bbe64c1986a356be2d6ba0a")
	})

	It("should ignore secret values file from the chart when no explicit option specified", func() {
		ctx := context.Background()
		bundleDir := ""
		secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{})

		bundle, err := NewBundle(ctx, bundleDir, helm_v3.Settings, nil, secretsManager, BundleOptions{
			SecretValueFiles: nil,
		})
		Expect(err).To(Succeed())

		ch := &chart.Chart{}
		bundle.ChartCreated(ch)

		files := []*chart.ChartExtenderBufferedFile{
			{
				Name: "secret-values.yaml",
				Data: []byte(`
testsecrets:
  testkey: 1000b45ee4272d14b30be2d20b5963f09e372fdfe761bf3913186938f4054d09ed0e
`),
			},
		}
		Expect(bundle.ChartLoaded(files)).To(Succeed())

		vals, err := bundle.MakeValues(map[string]interface{}{"one": 1, "two": 2})
		Expect(err).To(Succeed())

		Expect(vals).To(Equal(map[string]interface{}{"one": 1, "two": 2}))
	})

	It("should load from local FS secret values file specified with explicit option", func() {
		ctx := context.Background()
		bundleDir := ""

		secretsManager := secrets_manager.NewSecretsManager(secrets_manager.SecretsManagerOptions{})

		secretValuesFile, err := ioutil.TempFile("", "bundle-test-*.yaml")
		defer os.RemoveAll(secretValuesFile.Name())

		Expect(err).To(Succeed())
		Expect(os.WriteFile(secretValuesFile.Name(), []byte(`
testsecrets:
  testkey: 1000b45ee4272d14b30be2d20b5963f09e372fdfe761bf3913186938f4054d09ed0e
`), os.ModePerm)).To(Succeed())

		bundle, err := NewBundle(ctx, bundleDir, helm_v3.Settings, nil, secretsManager, BundleOptions{
			SecretValueFiles: []string{secretValuesFile.Name()},
		})
		Expect(err).To(Succeed())

		ch := &chart.Chart{}
		bundle.ChartCreated(ch)

		Expect(bundle.ChartLoaded(nil)).To(Succeed())

		vals, err := bundle.MakeValues(map[string]interface{}{"one": 1, "two": 2})
		Expect(err).To(Succeed())
		Expect(vals).To(Equal(map[string]interface{}{"one": 1, "two": 2, "testsecrets": map[string]interface{}{"testkey": "TOPSECRET"}}))
	})
})
