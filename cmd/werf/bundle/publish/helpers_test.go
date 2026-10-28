package publish

import (
	"context"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/secrets_manager"
	nelmcommon "github.com/werf/nelm/v2/pkg/common"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
)

var _ chartcommon.RuntimeData = (*fakeSecretsRuntimeData)(nil)

type fakeSecretsRuntimeData struct {
	defaultSecretValues map[string]interface{}
}

func (d *fakeSecretsRuntimeData) DecodeAndLoadSecrets(_ context.Context, _ []*nelmcommon.BufferedFile, _ *secrets_manager.SecretsManager, _ chartcommon.DecodeAndLoadSecretsOptions) error {
	return nil
}

func (d *fakeSecretsRuntimeData) GetDecryptedDefaultSecretValues() map[string]interface{} {
	return d.defaultSecretValues
}

func (d *fakeSecretsRuntimeData) GetDecryptedCustomSecretValues() map[string]interface{} {
	return nil
}

func (d *fakeSecretsRuntimeData) GetDecryptedSecretFilesData() map[string]string {
	return nil
}

func parseValues(data string) map[string]interface{} {
	vals := map[string]interface{}{}
	Expect(yaml.Unmarshal([]byte(data), &vals)).To(Succeed())

	return vals
}
